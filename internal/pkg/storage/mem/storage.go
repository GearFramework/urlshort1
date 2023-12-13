package mem

import (
	"context"
	"github.com/GearFramework/urlshort/internal/pkg"
	"sync"
)

type mapCodes map[string]Codes

type Codes struct {
	Code      string
	UserID    int
	IsDeleted bool
}

type mapURLs map[string]string

type Storage struct {
	sync.RWMutex
	codeByURL mapCodes
	urlByCode mapURLs
}

var lastUserID int = 0

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) InitStorage() error {
	s.codeByURL = make(mapCodes, 10)
	s.urlByCode = make(mapURLs, 10)
	return nil
}

func (s *Storage) Close() {
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
	if lastUserID < userID {
		lastUserID = userID
	}
	return nil
}

func (s *Storage) InsertBatch(ctx context.Context, userID int, batch [][]string) error {
	for _, pack := range batch {
		s.codeByURL[pack[0]] = Codes{UserID: userID, Code: pack[1]}
		s.urlByCode[pack[1]] = pack[0]
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

func (s *Storage) Count() int {
	return len(s.codeByURL)
}

func (s *Storage) Truncate() error {
	for url, code := range s.codeByURL {
		delete(s.codeByURL, url)
		delete(s.urlByCode, code.Code)
	}
	return nil
}

func (s *Storage) Ping() error {
	return nil
}
