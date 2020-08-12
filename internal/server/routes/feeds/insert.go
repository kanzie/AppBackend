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
	"fmt"
	"net/http"

	"github.com/SuperGreenLab/AppBackend/internal/data/db"
	"github.com/SuperGreenLab/AppBackend/internal/server/middlewares"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/rileyr/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"upper.io/db.v3/lib/sqlbuilder"
)

func fillUserEnd(sess sqlbuilder.Database, ueid uuid.UUID, collection string, all db.Objects, factory func() db.UserEndObject) {
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
					logrus.Errorln(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("x-sgl-token", tokenString)

				boxes := []db.Box{}
				sess.Select("*").From("boxes").Where("userid = ?", uid).And("deleted = ?", false).All(&boxes)
				fillUserEnd(sess, id, "boxes", db.Boxes(boxes), func() db.UserEndObject { return &db.UserEndBox{} })

				plants := []db.Plant{}
				sess.Select("*").From("plants").Where("userid = ?", uid).And("deleted = ?", false).All(&plants)
				fillUserEnd(sess, id, "plants", db.Plants(plants), func() db.UserEndObject { return &db.UserEndPlant{} })

				timelapses := []db.Timelapse{}
				sess.Select("*").From("timelapses").Where("userid = ?", uid).And("deleted = ?", false).All(&timelapses)
				fillUserEnd(sess, id, "timelapses", db.Timelapses(timelapses), func() db.UserEndObject { return &db.UserEndTimelapse{} })

				devices := []db.Device{}
				sess.Select("*").From("devices").Where("userid = ?", uid).And("deleted = ?", false).All(&devices)
				fillUserEnd(sess, id, "devices", db.Devices(devices), func() db.UserEndObject { return &db.UserEndDevice{} })

				feeds := []db.Feed{}
				sess.Select("*").From("feeds").Where("userid = ?", uid).And("deleted = ?", false).All(&feeds)
				fillUserEnd(sess, id, "feeds", db.Feeds(feeds), func() db.UserEndObject { return &db.UserEndFeed{} })

				feedEntries := []db.FeedEntry{}
				sess.Select("*").From("feedentries").Where("userid = ?", uid).And("deleted = ?", false).All(&feedEntries)
				fillUserEnd(sess, id, "feedentries", db.FeedEntries(feedEntries), func() db.UserEndObject { return &db.UserEndFeedEntry{} })

				feedMedias := []db.FeedMedia{}
				sess.Select("*").From("feedmedias").Where("userid = ?", uid).And("deleted = ?", false).All(&feedMedias)
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
		middlewares.CreateUserEndObjects("userend_boxes", func() db.UserEndObject { return &db.UserEndBox{} }),
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
		middlewares.CreateUserEndObjects("userend_plants", func() db.UserEndObject { return &db.UserEndPlant{} }),
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
		middlewares.CreateUserEndObjects("userend_timelapses", func() db.UserEndObject { return &db.UserEndTimelapse{} }),
	},
)

var createDeviceHandler = middlewares.InsertEndpoint(
	"devices",
	func() interface{} { return &db.Device{} },
	[]middleware.Middleware{middlewares.SetUserID},
	[]middleware.Middleware{
		middlewares.CreateUserEndObjects("userend_devices", func() db.UserEndObject { return &db.UserEndDevice{} }),
	},
)

var createFeedHandler = middlewares.InsertEndpoint(
	"feeds",
	func() interface{} { return &db.Feed{} },
	[]middleware.Middleware{middlewares.SetUserID},
	[]middleware.Middleware{
		middlewares.CreateUserEndObjects("userend_feeds", func() db.UserEndObject { return &db.UserEndFeed{} }),
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
		middlewares.CreateUserEndObjects("userend_feedentries", func() db.UserEndObject { return &db.UserEndFeedEntry{} }),
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
		middlewares.CreateUserEndObjects("userend_feedmedias", func() db.UserEndObject { return &db.UserEndFeedMedia{} }),
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
