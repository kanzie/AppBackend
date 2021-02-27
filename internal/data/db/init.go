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

package db

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
)

// Settings - db connection settings
var (
	Settings postgresql.ConnectionURL
	Sess     sqlbuilder.Database
)

// InitDB - initializes the Settings variable from config params
func InitDB() {
	Settings = postgresql.ConnectionURL{
		Host:     "postgres",
		Database: "sglapp",
		User:     "postgres",
		Password: viper.GetString("PGPassword"),
	}
	var err error
	Sess, err = postgresql.Open(Settings)
	if err != nil {
		logrus.Errorf("db.Open(): %q\n", err)
		logrus.Errorf("%q", err)
		return
	}
}
