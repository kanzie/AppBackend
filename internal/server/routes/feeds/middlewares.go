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
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/rileyr/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
)

type sessContextKey struct{}

func createDBSession(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var err error
		sess, err := postgresql.Open(settings)
		if err != nil {
			logrus.Errorf("db.Open(): %q\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer sess.Close()

		ctx := context.WithValue(r.Context(), sessContextKey{}, sess)
		fn(w, r.WithContext(ctx), p)
	}
}

type objectContextKey struct{}

func decodeJSON(fnObject func() interface{}) func(fn httprouter.Handle) httprouter.Handle {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			o := fnObject()
			err := decodeJSONBody(w, r, o)
			if err != nil {
				var mr *malformedRequest
				if errors.As(err, &mr) {
					logrus.Errorln(err.Error())
					http.Error(w, mr.msg, mr.status)
				} else {
					log.Println(err.Error())
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}
			ctx := context.WithValue(r.Context(), objectContextKey{}, o)
			fn(w, r.WithContext(ctx), p)
		}
	}
}

func setUserID(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		o := r.Context().Value(objectContextKey{}).(UserObject)
		uid := r.Context().Value(userIDContextKey{}).(uuid.UUID)

		o.SetUserID(uid)

		ctx := context.WithValue(r.Context(), objectContextKey{}, o)
		fn(w, r.WithContext(ctx), p)
	}
}

func checkAccessRight(collection, field string, optional bool, factory func() UserObject) middleware.Middleware {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			o := r.Context().Value(objectContextKey{}).(UserObject)
			uid := r.Context().Value(userIDContextKey{}).(uuid.UUID)
			sess := r.Context().Value(sessContextKey{}).(sqlbuilder.Database)

			if err := checkUserID(sess, uid, o, collection, field, optional, factory); err != nil {
				logrus.Errorln(err.Error())
				http.Error(w, "Parent is owned by another user", http.StatusUnauthorized)
				return
			}

			fn(w, r, p)
		}
	}
}

type insertedIDContextKey struct{}

func insertObject(collection string) func(fn httprouter.Handle) httprouter.Handle {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			o := r.Context().Value(objectContextKey{})
			sess := r.Context().Value(sessContextKey{}).(sqlbuilder.Database)
			col := sess.Collection(collection)
			id, err := col.Insert(o)
			if err != nil {
				logrus.Errorln(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), insertedIDContextKey{}, uuid.FromStringOrNil(string(id.([]uint8))))
			fn(w, r.WithContext(ctx), p)
		}
	}
}

type updatedIDContextKey struct{}

func updateObject(collection string) func(fn httprouter.Handle) httprouter.Handle {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			o := r.Context().Value(objectContextKey{}).(Object)
			sess := r.Context().Value(sessContextKey{}).(sqlbuilder.Database)
			col := sess.Collection(collection)
			err := col.Find(o.GetID()).Update(o)
			if err != nil {
				logrus.Errorln(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), updatedIDContextKey{}, o.GetID().UUID)
			fn(w, r.WithContext(ctx), p)
		}
	}
}

type userIDContextKey struct{}
type userEndIDContextKey struct{}

func jwtToken(fn httprouter.Handle) httprouter.Handle {
	hmacSampleSecret := []byte(viper.GetString("JWTSecret"))

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		authentication := r.Header.Get("Authentication")
		tokenString := strings.ReplaceAll(authentication, "Bearer ", "")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return hmacSampleSecret, nil
		})

		if err != nil {
			logrus.Errorln(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), userIDContextKey{}, uuid.FromStringOrNil(claims["userID"].(string)))
			if userEndID, ok := claims["userEndID"]; ok == true {
				ctx = context.WithValue(ctx, userEndIDContextKey{}, uuid.FromStringOrNil(userEndID.(string)))
			}
			fn(w, r.WithContext(ctx), p)
		} else {
			logrus.Errorln(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}
}

func userEndIDRequired(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ueid := r.Context().Value(userEndIDContextKey{})
		if ueid == nil {
			logrus.Errorln("Missing userEndID")
			http.Error(w, "Missing userEndID", http.StatusBadRequest)
			return
		}
		fn(w, r, p)
	}
}

func objectIDRequired(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		o := r.Context().Value(objectContextKey{}).(Object)
		if o.GetID().Valid == false {
			logrus.Errorln("Missing object's ID")
			http.Error(w, "Missing object's ID", http.StatusBadRequest)
			return
		}
		fn(w, r, p)
	}
}

func createUserEndObjects(collection string, factory func() UserEndObject) middleware.Middleware {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			sess := r.Context().Value(sessContextKey{}).(sqlbuilder.Database)
			uid := r.Context().Value(userIDContextKey{}).(uuid.UUID)
			ueid := r.Context().Value(userEndIDContextKey{}).(uuid.UUID)

			id := r.Context().Value(insertedIDContextKey{}).(uuid.UUID)

			uends := []UserEnd{}
			err := sess.Collection("userends").Find("userid", uid).All(&uends)
			if err != nil {
				logrus.Errorln(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, uend := range uends {
				ueo := factory()
				ueo.SetObjectID(id)
				ueo.SetUserEndID(uend.ID.UUID)
				if uend.ID.UUID == ueid {
					ueo.SetSent(true)
				} else {
					ueo.SetDirty(true)
				}
				sess.Collection(collection).Insert(ueo)
			}

			fn(w, r, p)
		}
	}
}

func updateUserEndObjects(collection, field string) middleware.Middleware {
	return func(fn httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			sess := r.Context().Value(sessContextKey{}).(sqlbuilder.Database)
			uid := r.Context().Value(userIDContextKey{}).(uuid.UUID)
			ueid := r.Context().Value(userEndIDContextKey{}).(uuid.UUID)

			id := r.Context().Value(updatedIDContextKey{}).(uuid.UUID)

			_, err := sess.Update(collection).Set("dirty", true).Where(field, id).And("userendid != ?", ueid).And("userendid in (select id from userends where userid = ?)", uid).Exec()
			if err != nil {
				logrus.Errorln(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fn(w, r, p)
		}
	}
}
