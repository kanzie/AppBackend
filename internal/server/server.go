/*
 * Copyright (C) 2019  SuperGreenLab <towelie@supergreenlab.com>
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

package server

import (
	"net/http"

	"github.com/SuperGreenLab/AppBackend/internal/data/storage"

	"github.com/SuperGreenLab/AppBackend/internal/data/db"
	"github.com/SuperGreenLab/AppBackend/internal/server/routes/feeds"
	"github.com/SuperGreenLab/AppBackend/internal/server/routes/metrics"
	"github.com/SuperGreenLab/AppBackend/internal/server/routes/users"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Start starts the server
func Start() {
	db.InitDB()
	storage.SetupBucket("feedmedias")

	router := httprouter.New()

	users.InitUsers(router)
	metrics.InitMetrics(router)
	feeds.InitFeeds(router)

	go func() { log.Fatal(http.ListenAndServe(":8080", router)) }()
}
