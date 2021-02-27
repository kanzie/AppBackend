/*
 * Copyright (C) 2021  SuperGreenLab <towelie@supergreenlab.com>
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

package db

import (
	"fmt"

	"github.com/gofrs/uuid"
)

func GetObjectsWithField(field string, value interface{}, collection string, obj interface{}) error {
	selector := Sess.Select("*").From(collection).Where(fmt.Sprintf("%s = ?", field), value)
	if err := selector.All(obj); err != nil {
		return err
	}

	return nil
}

func GetObjectWithField(field string, value interface{}, collection string, obj interface{}) error {
	selector := Sess.Select("*").From(collection).Where(fmt.Sprintf("%s = ?", field), value)
	if err := selector.One(obj); err != nil {
		return err
	}

	return nil
}

func GetObjectWithID(id uuid.UUID, collection string, obj interface{}) error {
	return GetObjectWithField("id", id, collection, obj)
}
