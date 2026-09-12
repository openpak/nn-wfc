package sake

import (
	"net/http"
	"strconv"
	"sync"
	"wwfc/common"
	"wwfc/logging"

	"github.com/logrusorgru/aurora/v3"
)

const (
	FileRequestDownload = iota
	FileRequestUpload
)

type FileRequest int

// Per-title file handlers, keyed by GameSpy game ID. Resolved lazily rather
// than at package init: the lookup needs game_list.tsv, and a working
// directory that does not have it should fail one request with a clear log
// line, not take down the whole server before main runs.
var fileDownloadHandlersOnce sync.Once
var fileDownloadHandlers map[int]func(string, http.ResponseWriter, *http.Request)
var fileUploadHandlersOnce sync.Once
var fileUploadHandlers map[int]func(string, http.ResponseWriter, *http.Request)

func fileDownloadHandler(id int) (func(string, http.ResponseWriter, *http.Request), bool) {
	fileDownloadHandlersOnce.Do(func() {
		fileDownloadHandlers = map[int]func(string, http.ResponseWriter, *http.Request){
			common.GetGameIDOrPanic("mariokartwii"): handleMarioKartWiiFileDownloadRequest,
		}
	})
	handler, ok := fileDownloadHandlers[id]
	return handler, ok
}

func fileUploadHandler(id int) (func(string, http.ResponseWriter, *http.Request), bool) {
	fileUploadHandlersOnce.Do(func() {
		fileUploadHandlers = map[int]func(string, http.ResponseWriter, *http.Request){
			common.GetGameIDOrPanic("mariokartwii"): handleMarioKartWiiFileUploadRequest,
		}
	})
	handler, ok := fileUploadHandlers[id]
	return handler, ok
}

func handleFileDownloadRequest(w http.ResponseWriter, r *http.Request) {
	moduleName := "SAKE:File:" + r.RemoteAddr

	gameIdString := r.URL.Query().Get("gameid")
	gameId, err := strconv.Atoi(gameIdString)
	if err != nil {
		logging.Error(moduleName, "Invalid GameSpy game ID:", aurora.Cyan(gameIdString))
		return
	}

	handler, handlerExists := fileDownloadHandler(gameId)
	if !handlerExists {
		logging.Warn(moduleName, "Unhandled file download request for GameSpy game ID:", aurora.Cyan(gameId))
		return
	}

	handler(moduleName, w, r)
}

func handleFileUploadRequest(w http.ResponseWriter, r *http.Request) {
	moduleName := "SAKE:File:" + r.RemoteAddr

	gameIdString := r.URL.Query().Get("gameid")
	gameId, err := strconv.Atoi(gameIdString)
	if err != nil {
		logging.Error(moduleName, "Invalid GameSpy game ID:", aurora.Cyan(gameIdString))
		return
	}

	handler, handlerExists := fileUploadHandler(gameId)
	if !handlerExists {
		logging.Warn(moduleName, "Unhandled file upload request for GameSpy game ID:", aurora.Cyan(gameId))
		return
	}

	handler(moduleName, w, r)
}
