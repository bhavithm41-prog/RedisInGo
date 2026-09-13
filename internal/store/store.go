package store

import (
	"errors"
	"sync"
	"time"
)

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

type Store struct {
	mu          sync.RWMutex
	data        map[string]any
	expirations map[string]time.Time
}

func New() *Store {
	return &Store{
		data:        make(map[string]any),
		expirations: make(map[string]time.Time),
	}
}

func (s *Store) isExpiredLocked(key string) bool {
	expiry, hasExpiry := s.expirations[key]
	if !hasExpiry {
		return false
	}
	return time.Now().After(expiry)
}

func (s *Store) expireIfNeededLocked(key string) {
	if s.isExpiredLocked(key) {
		delete(s.data, key)
		delete(s.expirations, key)
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	delete(s.expirations, key)
}

func (s *Store) Get(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	raw, exists := s.data[key]
	if !exists {
		return "", false, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", true, ErrWrongType
	}
	return value, true, nil
}

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		delete(s.expirations, key)
	}
	return exists
}

func (s *Store) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	_, exists := s.data[key]
	return exists
}

func (s *Store) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.data {
		s.expireIfNeededLocked(k)
	}
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Store) Expire(key string, seconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	_, exists := s.data[key]
	if !exists {
		return false
	}
	s.expirations[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	return true
}

func (s *Store) TTL(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	_, exists := s.data[key]
	if !exists {
		return -2
	}
	expiry, hasExpiry := s.expirations[key]
	if !hasExpiry {
		return -1
	}
	remaining := time.Until(expiry)
	if remaining < 0 {
		return -2
	}
	return int(remaining.Seconds())
}

func (s *Store) SetEx(key string, seconds int, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.expirations[key] = time.Now().Add(time.Duration(seconds) * time.Second)
}

func (s *Store) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	now := time.Now()
	for key, expiry := range s.expirations {
		if now.After(expiry) {
			delete(s.data, key)
			delete(s.expirations, key)
			removed++
		}
	}
	return removed
}

func (s *Store) LPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	list, err := s.getOrCreateList(key)
	if err != nil {
		return 0, err
	}
	for _, v := range values {
		list = append([]string{v}, list...)
	}
	s.data[key] = list
	return len(list), nil
}

func (s *Store) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	list, err := s.getOrCreateList(key)
	if err != nil {
		return 0, err
	}
	list = append(list, values...)
	s.data[key] = list
	return len(list), nil
}

func (s *Store) LPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	list, exists, err := s.getList(key)
	if err != nil {
		return "", false, err
	}
	if !exists || len(list) == 0 {
		return "", false, nil
	}
	value := list[0]
	list = list[1:]
	s.data[key] = list
	return value, true, nil
}

func (s *Store) RPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	list, exists, err := s.getList(key)
	if err != nil {
		return "", false, err
	}
	if !exists || len(list) == 0 {
		return "", false, nil
	}
	lastIdx := len(list) - 1
	value := list[lastIdx]
	list = list[:lastIdx]
	s.data[key] = list
	return value, true, nil
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	list, exists, err := s.getList(key)
	if err != nil {
		return nil, err
	}
	if !exists || len(list) == 0 {
		return []string{}, nil
	}
	length := len(list)
	start = normalizeIndex(start, length)
	stop = normalizeIndex(stop, length)
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}, nil
	}
	result := make([]string, stop-start+1)
	copy(result, list[start:stop+1])
	return result, nil
}

func (s *Store) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	set, err := s.getOrCreateSet(key)
	if err != nil {
		return 0, err
	}
	added := 0
	for _, m := range members {
		if _, exists := set[m]; !exists {
			set[m] = struct{}{}
			added++
		}
	}
	s.data[key] = set
	return added, nil
}

func (s *Store) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	set, exists, err := s.getSet(key)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}
	removed := 0
	for _, m := range members {
		if _, exists := set[m]; exists {
			delete(set, m)
			removed++
		}
	}
	s.data[key] = set
	return removed, nil
}

func (s *Store) SIsMember(key, member string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	set, exists, err := s.getSet(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	_, isMember := set[member]
	return isMember, nil
}

func (s *Store) SMembers(key string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	set, exists, err := s.getSet(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	members := make([]string, 0, len(set))
	for m := range set {
		members = append(members, m)
	}
	return members, nil
}

func (s *Store) HSet(key, field, value string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	hash, err := s.getOrCreateHash(key)
	if err != nil {
		return false, err
	}
	_, existed := hash[field]
	hash[field] = value
	s.data[key] = hash
	return !existed, nil
}

func (s *Store) HGet(key, field string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	hash, exists, err := s.getHash(key)
	if err != nil {
		return "", false, err
	}
	if !exists {
		return "", false, nil
	}
	value, fieldExists := hash[field]
	return value, fieldExists, nil
}

func (s *Store) HGetAll(key string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	hash, exists, err := s.getHash(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	result := make([]string, 0, len(hash)*2)
	for field, value := range hash {
		result = append(result, field, value)
	}
	return result, nil
}

func (s *Store) HDel(key, field string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expireIfNeededLocked(key)
	hash, exists, err := s.getHash(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	_, fieldExists := hash[field]
	if fieldExists {
		delete(hash, field)
		s.data[key] = hash
	}
	return fieldExists, nil
}

func (s *Store) getList(key string) ([]string, bool, error) {
	raw, exists := s.data[key]
	if !exists {
		return nil, false, nil
	}
	list, ok := raw.([]string)
	if !ok {
		return nil, true, ErrWrongType
	}
	return list, true, nil
}

func (s *Store) getOrCreateList(key string) ([]string, error) {
	list, exists, err := s.getList(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	return list, nil
}

func (s *Store) getSet(key string) (map[string]struct{}, bool, error) {
	raw, exists := s.data[key]
	if !exists {
		return nil, false, nil
	}
	set, ok := raw.(map[string]struct{})
	if !ok {
		return nil, true, ErrWrongType
	}
	return set, true, nil
}

func (s *Store) getOrCreateSet(key string) (map[string]struct{}, error) {
	set, exists, err := s.getSet(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make(map[string]struct{}), nil
	}
	return set, nil
}

func (s *Store) getHash(key string) (map[string]string, bool, error) {
	raw, exists := s.data[key]
	if !exists {
		return nil, false, nil
	}
	hash, ok := raw.(map[string]string)
	if !ok {
		return nil, true, ErrWrongType
	}
	return hash, true, nil
}

func (s *Store) getOrCreateHash(key string) (map[string]string, error) {
	hash, exists, err := s.getHash(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make(map[string]string), nil
	}
	return hash, nil
}

func normalizeIndex(idx, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}
