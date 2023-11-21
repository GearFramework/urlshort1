package file

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/GearFramework/urlshort/internal/pkg"
	"github.com/GearFramework/urlshort/internal/pkg/logger"
	"io"
	"log"
	"os"
	"sync"
)

type Codes struct {
	Code      string `json:"code"`
	UserID    int    `json:"user_id"`
	IsDeleted bool   `json:"is_deleted"`
}

type Storage struct {
	sync.RWMutex
	Config       *StorageConfig
	codeByURL    map[string]Codes
	urlByCode    map[string]string
	flushCounter int
}

var lastUserID int = 0

func NewStorage(config *StorageConfig) *Storage {
	return &Storage{
		Config: config,
	}
}

func (s *Storage) InitStorage() error {
	s.codeByURL = make(map[string]Codes, s.Config.FlushPerItems)
	s.urlByCode = make(map[string]string, s.Config.FlushPerItems)
	s.flushCounter = s.Config.FlushPerItems
	err := s.loadShortlyURLs()
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (s *Storage) loadShortlyURLs() error {
	file, err := os.OpenFile(s.Config.StorageFilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(&s.codeByURL); err != nil {
		return err
	}
	s.urlByCode = make(map[string]string, s.Config.FlushPerItems)
	for url, data := range s.codeByURL {
		s.urlByCode[data.Code] = url
		if lastUserID < data.UserID {
			lastUserID = data.UserID
		}
	}
	s.flushCounter = s.Count() + s.Config.FlushPerItems
	return nil
}

func (s *Storage) Close() {
	if err := s.flush(); err != nil {
		logger.Log.Error(err.Error())
	}
}

func (s *Storage) GetCode(ctx context.Context, url string) (string, bool) {
	data, ok := s.codeByURL[url]
	if ok {
		return data.Code, ok
	}
	return "", ok
}

func (s *Storage) GetCodeBatch(ctx context.Context, batch []string) map[string]string {
	codes := map[string]string{}
	for _, url := range batch {
		if _, ok := s.codeByURL[url]; ok {
			codes[url] = s.codeByURL[url].Code
		}
	}
	return codes
}

func (s *Storage) GetURL(ctx context.Context, code string) (pkg.ShortURL, bool) {
	url, ok := s.urlByCode[code]
	short := pkg.ShortURL{}
	if ok {
		short.URL = url
		short.IsDeleted = s.codeByURL[url].IsDeleted
	}
	return short, ok
}

func (s *Storage) GetMaxUserID(ctx context.Context) (int, error) {
	return lastUserID, nil
}

func (s *Storage) GetUserURLs(ctx context.Context, userID int) []pkg.UserURL {
	userURLs := []pkg.UserURL{}
	for url, userShortURL := range s.codeByURL {
		if userShortURL.UserID == userID {
			userURLs = append(userURLs, pkg.UserURL{Code: userShortURL.Code, URL: url})
		}
	}
	return userURLs
}

func (s *Storage) Insert(ctx context.Context, userID int, url, code string) error {
	s.codeByURL[url] = Codes{UserID: userID, Code: code}
	s.urlByCode[code] = url
	var err error
	if s.mustFlush() {
		err = s.flush()
		s.flushCounter += s.Config.FlushPerItems
	}
	if lastUserID < userID {
		lastUserID = userID
	}
	return err
}

func (s *Storage) InsertBatch(ctx context.Context, userID int, batch [][]string) error {
	for _, pack := range batch {
		s.codeByURL[pack[0]] = Codes{UserID: userID, Code: pack[1]}
		s.urlByCode[pack[1]] = pack[0]
	}
	if s.mustFlush() {
		if err := s.flush(); err != nil {
			logger.Log.Warn(err.Error())
		}
		s.flushCounter += s.Config.FlushPerItems
	}
	if lastUserID < userID {
		lastUserID = userID
	}
	return nil
}

func (s *Storage) DeleteBatch(ctx context.Context, userID int, batch []string) {
	s.Lock()
	defer s.Unlock()
	for _, code := range batch {
		url, ok := s.urlByCode[code]
		if ok && s.codeByURL[url].UserID == userID {
			short := s.codeByURL[url]
			short.IsDeleted = true
			s.codeByURL[url] = short
		}
	}
}

func (s *Storage) mustFlush() bool {
	return s.Count() == s.flushCounter
}

func (s *Storage) Count() int {
	return len(s.codeByURL)
}

func (s *Storage) flush() error {
	if s.codeByURL == nil {
		return nil
	}
	data, err := json.Marshal(&s.codeByURL)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Config.StorageFilePath, data, 0666)
}

func (s *Storage) Truncate() error {
	for url, code := range s.codeByURL {
		delete(s.codeByURL, url)
		delete(s.urlByCode, code.Code)
	}
	return s.flush()
}

func (s *Storage) Ping() error {
	_, err := os.Stat(s.Config.StorageFilePath)
	if os.IsNotExist(err) {
		fd, err := os.OpenFile(s.Config.StorageFilePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		return fd.Close()
	}
	return nil
}
