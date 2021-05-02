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
	cmiddlewares "github.com/SuperGreenLab/AppBackend/internal/server/middlewares"
	"github.com/SuperGreenLab/AppBackend/internal/server/routes/feeds/explorer"
	fmiddlewares "github.com/SuperGreenLab/AppBackend/internal/server/routes/feeds/middlewares"
	"github.com/julienschmidt/httprouter"
)

// Init -
func Init(router *httprouter.Router) {
	//anon := cmiddlewares.AnonStack()
	auth := cmiddlewares.AuthStack()
	optionalAuth := cmiddlewares.OptionalAuthStack()
	authWithUserEndID := fmiddlewares.AuthStackWithUserEnd()

	router.POST("/userend", auth.Wrap(createUserEndHandler))
	router.POST("/plantsharing", auth.Wrap(createPlantSharingHandler))

	router.POST("/box", auth.Wrap(createBoxHandler))
	router.POST("/plant", auth.Wrap(createPlantHandler))
	router.POST("/timelapse", auth.Wrap(createTimelapseHandler))
	router.POST("/timelapseframe", auth.Wrap(createTimelapseFrameHandler))
	router.POST("/device", auth.Wrap(createDeviceHandler))
	router.POST("/feed", auth.Wrap(createFeedHandler))
	router.POST("/feedEntry", auth.Wrap(createFeedEntryHandler))
	router.POST("/feedMedia", auth.Wrap(createFeedMediaHandler))
	router.POST("/comment", auth.Wrap(createCommentHandler))
	router.POST("/like", auth.Wrap(createLikeHandler))
	router.POST("/report", auth.Wrap(createReportHandler))
	router.POST("/bookmark", auth.Wrap(createBookmarkHandler))
	router.POST("/follow", auth.Wrap(createFollowHandler))
	router.POST("/linkbookmark", auth.Wrap(createLinkBookmarkHandler))

	router.PUT("/box", auth.Wrap(updateBoxHandler))
	router.PUT("/plant", auth.Wrap(updatePlantHandler))
	router.PUT("/timelapse", auth.Wrap(updateTimelapseHandler))
	router.PUT("/device", auth.Wrap(updateDeviceHandler))
	router.PUT("/feed", auth.Wrap(updateFeedHandler))
	router.PUT("/feedEntry", auth.Wrap(updateFeedEntryHandler))
	router.PUT("/feedMedia", auth.Wrap(updateFeedMediaHandler))
	router.PUT("/userend", auth.Wrap(updateUserEndHandler))

	router.POST("/deletes", auth.Wrap(deletesHandler))

	router.POST("/feedMediaUploadURL", auth.Wrap(feedMediaUploadURLHandler))
	router.POST("/timelapseUploadURL", auth.Wrap(timelapseUploadURLHandler))

	router.GET("/syncBoxes", authWithUserEndID.Wrap(syncBoxesHandler))
	router.GET("/syncPlants", authWithUserEndID.Wrap(syncPlantsHandler))
	router.GET("/syncTimelapses", authWithUserEndID.Wrap(syncTimelapsesHandler))
	router.GET("/syncDevices", authWithUserEndID.Wrap(syncDevicesHandler))
	router.GET("/syncFeeds", authWithUserEndID.Wrap(syncFeedsHandler))
	router.GET("/syncFeedEntries", authWithUserEndID.Wrap(syncFeedEntriesHandler))
	router.GET("/syncFeedMedias", authWithUserEndID.Wrap(syncFeedMediasHandler))

	router.POST("/box/:id/sync", authWithUserEndID.Wrap(syncedBoxHandler))
	router.POST("/plant/:id/sync", authWithUserEndID.Wrap(syncedPlantHandler))
	router.POST("/timelapse/:id/sync", authWithUserEndID.Wrap(syncedTimelapseHandler))
	router.POST("/device/:id/sync", authWithUserEndID.Wrap(syncedDeviceHandler))
	router.POST("/feed/:id/sync", authWithUserEndID.Wrap(syncedFeedHandler))
	router.POST("/feedEntry/:id/sync", authWithUserEndID.Wrap(syncedFeedEntryHandler))
	router.POST("/feedMedia/:id/sync", authWithUserEndID.Wrap(syncedFeedMediaHandler))

	router.POST("/plant/:id/archive", authWithUserEndID.Wrap(archivePlantHandler))

	router.GET("/plants", auth.Wrap(selectPlants))
	router.GET("/plant/:id", auth.Wrap(selectPlant))
	router.GET("/feedEntries", auth.Wrap(selectFeedEntries))
	router.GET("/feedEntry/:id", auth.Wrap(selectFeedEntry))
	router.GET("/feedEntry/:id/comments", optionalAuth.Wrap(selectFeedEntryComments))
	router.GET("/feedEntry/:id/comments/count", optionalAuth.Wrap(countFeedEntryComments))
	router.GET("/feedEntry/:id/social", optionalAuth.Wrap(selectFeedEntrySocial))
	router.GET("/comment/:id", optionalAuth.Wrap(selectComment))
	router.GET("/feedMedias", auth.Wrap(selectFeedMedias))
	router.GET("/feedMedia/:id", auth.Wrap(selectFeedMedia))
	router.GET("/feeds", auth.Wrap(selectFeeds))
	router.GET("/feed/:id", auth.Wrap(selectFeed))
	router.GET("/boxes", auth.Wrap(selectBoxes))
	router.GET("/box/:id", auth.Wrap(selectBox))
	router.GET("/devices", auth.Wrap(selectDevices))
	router.GET("/device/:id", auth.Wrap(selectDevice))
	router.GET("/device/:id/params", auth.Wrap(selectDeviceParams))
	router.GET("/bookmarks", auth.Wrap(selectBookmarks))
	router.GET("/bookmark/:id", auth.Wrap(selectBookmark))
	router.GET("/timelapses", auth.Wrap(selectTimelapses))
	router.GET("/timelapse/:id", auth.Wrap(selectTimelapse))

	explorer.Init(router)
}
