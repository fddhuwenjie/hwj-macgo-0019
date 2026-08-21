package domain

import "time"

type FailureStreak struct {
	BaseAggregate
	categoryID      string
	count           int
	since           time.Time
	lastFailureAt   time.Time
	currentWindowID string
}

func NewFailureStreak(id, categoryID, windowID string, now time.Time) (*FailureStreak, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	if categoryID == "" || windowID == "" {
		return nil, NewError(ErrorKindInvalidArgument, "NewFailureStreak", ErrInvalidInvariant)
	}
	s := &FailureStreak{
		categoryID:      categoryID,
		currentWindowID: windowID,
		since:           now,
	}
	s.setID(id)
	s.setVersion(1)
	s.setCreatedAt(now)
	s.setUpdatedAt(now)
	return s, nil
}

func (s *FailureStreak) CategoryID() string {
	if s == nil {
		return ""
	}
	return s.categoryID
}

func (s *FailureStreak) Count() int {
	if s == nil {
		return 0
	}
	return s.count
}

func (s *FailureStreak) Since() time.Time {
	if s == nil {
		return time.Time{}
	}
	return s.since
}

func (s *FailureStreak) LastFailureAt() time.Time {
	if s == nil {
		return time.Time{}
	}
	return s.lastFailureAt
}

func (s *FailureStreak) CurrentWindowID() string {
	if s == nil {
		return ""
	}
	return s.currentWindowID
}

func (s *FailureStreak) RecordFailure(now time.Time) (int, error) {
	if s == nil {
		return 0, ErrNilArgument
	}
	s.count++
	s.lastFailureAt = now
	s.bumpVersion(now)
	return s.count, nil
}

func (s *FailureStreak) RecordSuccess(now time.Time) error {
	if s == nil {
		return ErrNilArgument
	}
	s.count = 0
	s.currentWindowID = ""
	s.lastFailureAt = time.Time{}
	s.bumpVersion(now)
	return nil
}

func (s *FailureStreak) RotateToWindow(windowID string, now time.Time) error {
	if s == nil {
		return ErrNilArgument
	}
	if windowID == "" {
		return NewError(ErrorKindInvalidArgument, "FailureStreak.RotateToWindow", ErrInvalidInvariant)
	}
	if windowID == s.currentWindowID {
		return nil
	}
	// 切换窗口必须同时清理连续失败计数，保证告警阈值按窗口隔离。
	// count 不归零会使 ValidateBackoffLimit 带着上个窗口的失败次数判定当前窗口，
	// 导致新窗口未发生失败即触发 ErrSuspensionRequired。
	s.count = 0
	s.currentWindowID = windowID
	s.lastFailureAt = time.Time{}
	s.since = now
	s.bumpVersion(now)
	return nil
}

func (s *FailureStreak) Clone() *FailureStreak {
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}
