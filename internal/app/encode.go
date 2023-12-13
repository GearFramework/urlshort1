package app

import (
	"context"
	"fmt"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"runtime"
)

const (
	alphabet    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	lenAlpha    = len(alphabet)
	defShortLen = 8
)

func (app *ShortApp) EncodeURL(ctx context.Context, userID int, url string) (string, bool) {
	app.Store.Lock()
	defer app.Store.Unlock()
	code, exists := app.Store.GetCode(ctx, url)
	if !exists {
		code = app.getRandomString(defShortLen)
		if err := app.Store.Insert(ctx, userID, url, code); err != nil {
			logger.Log.Error(err.Error())
		}
	}
	return fmt.Sprintf("%s/%s", app.Conf.ShortURLHost, code), exists
}

func (app *ShortApp) BatchEncodeURL(ctx context.Context, userID int, batch []pkg.BatchURLs) []pkg.ResultBatchShort {
	app.Store.Lock()
	defer app.Store.Unlock()
	res := []pkg.ResultBatchShort{}
	trc, urls := transformBatchByCorrelation(batch)
	for _, chunkURLs := range chunkingURLs(urls) {
		existCodes := app.Store.GetCodeBatch(ctx, chunkURLs)
		for existURL, existCode := range existCodes {
			res = append(res, pkg.ResultBatchShort{
				CorrelationID: trc[existURL],
				ShortURL:      fmt.Sprintf("%s/%s", app.Conf.ShortURLHost, existCode),
			})
		}
		if len(chunkURLs) == len(existCodes) {
			continue
		}
		notExistCodes := getNotExists(chunkURLs, existCodes)
		newShortURLs, pack := app.prepareNotExistsShortURLs(trc, notExistCodes)
		if err := app.Store.InsertBatch(ctx, userID, pack); err != nil {
			logger.Log.Error(err.Error())
			continue
		}
		res = append(res, newShortURLs...)
	}
	return res
}

func (app *ShortApp) prepareNotExistsShortURLs(
	trc map[string]string,
	notExistCodes []string,
) ([]pkg.ResultBatchShort, [][]string) {
	res := []pkg.ResultBatchShort{}
	pack := [][]string{}
	for _, url := range notExistCodes {
		code := app.getRandomString(defShortLen)
		res = append(res, pkg.ResultBatchShort{
			CorrelationID: trc[url],
			ShortURL:      fmt.Sprintf("%s/%s", app.Conf.ShortURLHost, code),
		})
		pack = append(pack, []string{url, code})
	}
	return res, pack
}

func transformBatchByCorrelation(batch []pkg.BatchURLs) (map[string]string, []string) {
	trc := map[string]string{}
	urls := []string{}
	for _, packet := range batch {
		trc[packet.OriginalURL] = packet.CorrelationID
		urls = append(urls, packet.OriginalURL)
	}
	return trc, urls
}

func chunkingURLs(urls []string) [][]string {
	var chunks [][]string
	count := len(urls)
	numCPU := runtime.NumCPU()
	var chunkSize int
	if count <= numCPU {
		chunkSize = count
	} else {
		chunkSize = (count + numCPU - 1) / numCPU
	}
	for i := 0; i < count; i += chunkSize {
		end := i + chunkSize
		if end > count {
			end = count
		}
		chunks = append(chunks, urls[i:end])
	}
	return chunks
}

func getNotExists(chunkURLs []string, exists map[string]string) []string {
	if len(chunkURLs) == len(exists) {
		return []string{}
	}
	notExists := []string{}
	for _, url := range chunkURLs {
		if _, ok := exists[url]; ok {
			continue
		}
		notExists = append(notExists, url)
	}
	return notExists
}
