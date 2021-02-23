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

package prometheus

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPTiming struct {
	router *httprouter.Router
}

func (ht *HTTPTiming) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
	ht.router.ServeHTTP(w, r)
	timer.ObserveDuration()
}

func NewHTTPTiming(router *httprouter.Router) *HTTPTiming {
	return &HTTPTiming{router}
}

func Init() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
}
