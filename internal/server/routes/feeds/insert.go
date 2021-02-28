/*
 * Copyright (C) 2020  SuperGreenLab <towelie@supergreenlab.com>
 * Author: Constantin Clauzel <constantin.clauzel@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package feeds

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SuperGreenLab/AppBackend/internal/data/db"
	"github.com/SuperGreenLab/AppBackend/internal/server/middlewares"
	fmiddlewares "github.com/SuperGreenLab/AppBackend/internal/server/routes/feeds/middlewares"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rileyr/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	udb "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

func fillUserEnd(sess sqlbuilder.Database, ueid uuid.UUID, collection string, all db.Objects, factory func() db.UserEndObject) {
	// TODO batch insert, or insert select below
	all.Each(func(a db.Object) {
		ueo := factory()
		ueo.SetUserEndID(ueid)
		ueo.SetObjectID(a.GetID().UUID)
		ueo.SetDirty(true)
		sess.Collection(fmt.Sprintf("userend_%s", collection)).Insert(ueo)
	})
}

var createUserEndHandler = middlewares.InsertEndpoint(
	"userends",
	func() interface{} { return &db.UserEnd{} },
	[]middleware.Middleware{middlewares.SetUserID},
	[]middleware.Middleware{
		func(fn httprouter.Handle) httprouter.Handle {
			return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
				hmacSampleSecret := []byte(viper.GetString("JWTSecret"))
				sess := r.Context().Value(middlewares.SessContextKey{}).(sqlbuilder.Database)
				id := r.Context().Value(middlewares.InsertedIDContextKey{}).(uuid.UUID)
				uid := r.Context().Value(middlewares.UserIDContextKey{}).(uuid.UUID)

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"userID":    uid.String(),
					"userEndID": id.String(),
				})
				tokenString, err := token.SignedString(hmacSampleSecret)
				if err != nil {
					logrus.Errorf("token.SignedString in createUserEndHandler %q - userID: %s userEndID: %s", err, uid, id)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("x-sgl-token", tokenString)

				boxes := []db.Box{}
				err = sess.Select("*").From("boxes").Where("userid = ?", uid).And("deleted = ?", false).All(&boxes)
				if err != nil {
					logrus.Errorf("sess.Select.From('boxes') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "boxes", db.Boxes(boxes), func() db.UserEndObject { return &db.UserEndBox{} })

				plants := []db.Plant{}
				err = sess.Select("*").From("plants").Where("userid = ?", uid).And("deleted = ?", false).And("archived = ?", false).All(&plants)
				if err != nil {
					logrus.Errorf("sess.Select.From('plants') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "plants", db.Plants(plants), func() db.UserEndObject { return &db.UserEndPlant{} })

				timelapses := []db.Timelapse{}
				err = sess.Select("*").From("timelapses").Where("userid = ?", uid).And("deleted = ?", false).And("(select archived from plants where plants.id = timelapses.plantid) = ?", false).All(&timelapses)
				if err != nil {
					logrus.Errorf("sess.Select.From('timelapses') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "timelapses", db.Timelapses(timelapses), func() db.UserEndObject { return &db.UserEndTimelapse{} })

				devices := []db.Device{}
				err = sess.Select("*").From("devices").Where("userid = ?", uid).And("deleted = ?", false).All(&devices)
				if err != nil {
					logrus.Errorf("sess.Select.From('devices') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "devices", db.Devices(devices), func() db.UserEndObject { return &db.UserEndDevice{} })

				feeds := []db.Feed{}
				err = sess.Select("*").From("feeds").Where("userid = ?", uid).And("deleted = ?", false).And(
					udb.Or(
						udb.Raw("not exists(select id from plants where plants.feedid = feeds.id)"),
						udb.Raw("(select archived from plants where plants.feedid = feeds.id) = ?", false)),
				).All(&feeds)
				if err != nil {
					logrus.Errorf("sess.Select.From('feeds') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "feeds", db.Feeds(feeds), func() db.UserEndObject { return &db.UserEndFeed{} })

				feedEntries := []db.FeedEntry{}
				err = sess.Select("*").From("feedentries").Where("userid = ?", uid).And("deleted = ?", false).And(
					udb.Or(
						udb.Raw("not exists(select id from plants where plants.feedid = feedentries.feedid)"),
						udb.Raw("(select archived from plants where plants.feedid = feedentries.feedid) = ?", false)),
				).All(&feedEntries)
				if err != nil {
					logrus.Errorf("sess.Select.From('feedentries') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "feedentries", db.FeedEntries(feedEntries), func() db.UserEndObject { return &db.UserEndFeedEntry{} })

				feedMedias := []db.FeedMedia{}
				err = sess.Select("*").From("feedmedias").Where("userid = ?", uid).And("deleted = ?", false).And(
					udb.Or(
						udb.Raw("not exists(select id from plants where plants.feedid = (select feedid from feedentries where feedmedias.feedentryid = feedentries.id))"),
						udb.Raw("(select archived from plants where plants.feedid = (select feedid from feedentries where feedmedias.feedentryid = feedentries.id)) = ?", false)),
				).All(&feedMedias)
				if err != nil {
					logrus.Errorf("sess.Select.From('feedmedias') in createUserEndHandler %q - uid: %s", err, uid)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				fillUserEnd(sess, id, "feedmedias", db.FeedMedias(feedMedias), func() db.UserEndObject { return &db.UserEndFeedMedia{} })

				fn(w, r, p)
			}
		},
	},
)

var createBoxHandler = middlewares.InsertEndpoint(
	"boxes",
	func() interface{} { return &db.Box{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("devices", "DeviceID", true, func() db.UserObject { return &db.Device{} }),
	},
	[]middleware.Middleware{
		fmiddlewares.CreateUserEndObjects("userend_boxes", func() db.UserEndObject { return &db.UserEndBox{} }),
	},
)

var createPlantHandler = middlewares.InsertEndpoint(
	"plants",
	func() interface{} { return &db.Plant{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("boxes", "BoxID", false, func() db.UserObject { return &db.Box{} }),
	},
	[]middleware.Middleware{
		fmiddlewares.CreateUserEndObjects("userend_plants", func() db.UserEndObject { return &db.UserEndPlant{} }),
	},
)

var createTimelapseHandler = middlewares.InsertEndpoint(
	"timelapses",
	func() interface{} { return &db.Timelapse{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("plants", "PlantID", false, func() db.UserObject { return &db.Plant{} }),
	},
	[]middleware.Middleware{
		fmiddlewares.CheckPlantArchivedForTimelapse,
		fmiddlewares.CreateUserEndObjects("userend_timelapses", func() db.UserEndObject { return &db.UserEndTimelapse{} }),
	},
)

var createDeviceHandler = middlewares.InsertEndpoint(
	"devices",
	func() interface{} { return &db.Device{} },
	[]middleware.Middleware{middlewares.SetUserID},
	[]middleware.Middleware{
		fmiddlewares.CreateUserEndObjects("userend_devices", func() db.UserEndObject { return &db.UserEndDevice{} }),
	},
)

var createFeedHandler = middlewares.InsertEndpoint(
	"feeds",
	func() interface{} { return &db.Feed{} },
	[]middleware.Middleware{middlewares.SetUserID},
	[]middleware.Middleware{
		fmiddlewares.CheckPlantArchivedForFeed,
		fmiddlewares.CreateUserEndObjects("userend_feeds", func() db.UserEndObject { return &db.UserEndFeed{} }),
	},
)

var createFeedEntryHandler = middlewares.InsertEndpoint(
	"feedentries",
	func() interface{} { return &db.FeedEntry{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("feeds", "FeedID", false, func() db.UserObject { return &db.Feed{} }),
	},
	[]middleware.Middleware{
		fmiddlewares.CheckPlantArchivedForFeedEntry,
		fmiddlewares.CreateUserEndObjects("userend_feedentries", func() db.UserEndObject { return &db.UserEndFeedEntry{} }),
	},
)

var createFeedMediaHandler = middlewares.InsertEndpoint(
	"feedmedias",
	func() interface{} { return &db.FeedMedia{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("feedentries", "FeedEntryID", false, func() db.UserObject { return &db.FeedEntry{} }),
	},
	[]middleware.Middleware{
		fmiddlewares.CheckPlantArchivedForFeedMedia,
		fmiddlewares.CreateUserEndObjects("userend_feedmedias", func() db.UserEndObject { return &db.UserEndFeedMedia{} }),
	},
)

var createPlantSharingHandler = middlewares.InsertEndpoint(
	"plantsharings",
	func() interface{} { return &db.PlantSharing{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
		middlewares.CheckAccessRight("feedentries", "FeedEntryID", false, func() db.UserObject { return &db.FeedEntry{} }),
	},
	nil,
)

var createCommentHandler = middlewares.InsertEndpoint(
	"comments",
	func() interface{} { return &db.Comment{} },
	[]middleware.Middleware{
		middlewares.SetUserID,
	},
	nil,
)

func deleteLikeIfExists(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sess := r.Context().Value(middlewares.SessContextKey{}).(sqlbuilder.Database)
		uid := r.Context().Value(middlewares.UserIDContextKey{}).(uuid.UUID)
		l := r.Context().Value(middlewares.ObjectContextKey{}).(*db.Like)

		var like db.Like
		err := sess.Collection("likes").Find().Where("userid = ?", uid).And(udb.Or(udb.Raw("commentid = ?", l.CommentID), udb.Raw("feedentryid = ?", l.FeedEntryID))).One(&like)
		if err == nil {
			err := sess.Collection("likes").Find().Where("id = ?", like.ID).Delete()
			if err != nil {
				logrus.Errorf("sess.Collection('likes') in deleteLikeIfExists %q %+v", err, like)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), middlewares.ObjectContextKey{}, like)
			ctx = context.WithValue(ctx, middlewares.InsertedIDContextKey{}, like.ID.UUID)
			middlewares.OutputObjectID(w, r.WithContext(ctx), p)
		} else {
			logrus.Infof("sess.Collection('likes') in deleteLikeIfExists %q", err)
			fn(w, r, p)
		}
	}
}

var createLikeHandler = middlewares.InsertEndpoint(
	"likes",
	func() interface{} { return &db.Like{} },
	[]middleware.Middleware{
		deleteLikeIfExists,
		middlewares.SetUserID,
	},
	nil,
)

func ignoreReportIfExists(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sess := r.Context().Value(middlewares.SessContextKey{}).(sqlbuilder.Database)
		uid := r.Context().Value(middlewares.UserIDContextKey{}).(uuid.UUID)
		re := r.Context().Value(middlewares.ObjectContextKey{}).(*db.Report)

		var report db.Report
		err := sess.Collection("reports").Find().Where("userid = ?", uid).And(udb.Or(udb.Raw("plantid = ?", re.PlantID), udb.Raw("commentid = ?", re.CommentID), udb.Raw("feedentryid = ?", re.FeedEntryID))).One(&report)
		if err == nil {
			ctx := context.WithValue(r.Context(), middlewares.ObjectContextKey{}, report)
			ctx = context.WithValue(ctx, middlewares.InsertedIDContextKey{}, report.ID.UUID)
			middlewares.OutputObjectID(w, r.WithContext(ctx), p)
		} else {
			logrus.Println(err)
			fn(w, r, p)
		}
	}
}

var createReportHandler = middlewares.InsertEndpoint(
	"reports",
	func() interface{} { return &db.Report{} },
	[]middleware.Middleware{
		ignoreReportIfExists,
		middlewares.SetUserID,
	},
	nil,
)

func deleteBookmarkIfExists(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		sess := r.Context().Value(middlewares.SessContextKey{}).(sqlbuilder.Database)
		uid := r.Context().Value(middlewares.UserIDContextKey{}).(uuid.UUID)
		b := r.Context().Value(middlewares.ObjectContextKey{}).(*db.Bookmark)

		var bookmark db.Bookmark
		err := sess.Collection("bookmarks").Find().Where("userid = ?", uid).And("feedentryid = ?", b.FeedEntryID).One(&bookmark)
		if err == nil {
			err := sess.Collection("bookmarks").Find().Where("id = ?", bookmark.ID).Delete()
			if err != nil {
				logrus.Errorf("sess.Collection('bookmarks') in deleteBookmarkIfExists %q - %+v", err, bookmark)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), middlewares.ObjectContextKey{}, bookmark)
			ctx = context.WithValue(ctx, middlewares.InsertedIDContextKey{}, bookmark.ID.UUID)
			middlewares.OutputObjectID(w, r.WithContext(ctx), p)
		} else {
			logrus.Infof("sess.Collection('bookmarks') in deleteLikeIfExists %q", err)
			fn(w, r, p)
		}
	}
}

var createBookmarkHandler = middlewares.InsertEndpoint(
	"bookmarks",
	func() interface{} { return &db.Bookmark{} },
	[]middleware.Middleware{
		deleteBookmarkIfExists,
		middlewares.SetUserID,
	},
	nil,
)
