package candidate_test

import (
	"context"
	"testing"

	"pemira-api/internal/candidate"
)

// Mock repository for testing
type mockCandidateRepo struct {
	candidates map[int64]*candidate.Candidate
	nextID     int64
}

func newMockRepo() *mockCandidateRepo {
	return &mockCandidateRepo{
		candidates: make(map[int64]*candidate.Candidate),
		nextID:     1,
	}
}

func (m *mockCandidateRepo) ListByElection(ctx context.Context, electionID int64, filter candidate.Filter) ([]candidate.Candidate, int64, error) {
	var results []candidate.Candidate
	for _, c := range m.candidates {
		if c.ElectionID == electionID {
			if filter.Status != nil && c.Status != *filter.Status {
				continue
			}
			results = append(results, *c)
		}
	}
	return results, int64(len(results)), nil
}

func (m *mockCandidateRepo) GetByID(ctx context.Context, electionID, candidateID int64) (*candidate.Candidate, error) {
	c, exists := m.candidates[candidateID]
	if !exists || c.ElectionID != electionID {
		return nil, candidate.ErrCandidateNotFound
	}
	return c, nil
}

func (m *mockCandidateRepo) Create(ctx context.Context, c *candidate.Candidate) (*candidate.Candidate, error) {
	c.ID = m.nextID
	m.nextID++
	m.candidates[c.ID] = c
	return c, nil
}

func (m *mockCandidateRepo) Update(ctx context.Context, electionID, candidateID int64, c *candidate.Candidate) (*candidate.Candidate, error) {
	existing, exists := m.candidates[candidateID]
	if !exists || existing.ElectionID != electionID {
		return nil, candidate.ErrCandidateNotFound
	}
	c.ID = candidateID
	m.candidates[candidateID] = c
	return c, nil
}

func (m *mockCandidateRepo) Delete(ctx context.Context, electionID, candidateID int64) error {
	existing, exists := m.candidates[candidateID]
	if !exists || existing.ElectionID != electionID {
		return candidate.ErrCandidateNotFound
	}
	delete(m.candidates, candidateID)
	return nil
}

func (m *mockCandidateRepo) UpdateStatus(ctx context.Context, electionID, candidateID int64, status candidate.CandidateStatus) error {
	c, exists := m.candidates[candidateID]
	if !exists || c.ElectionID != electionID {
		return candidate.ErrCandidateNotFound
	}
	c.Status = status
	return nil
}

func (m *mockCandidateRepo) CheckNumberExists(ctx context.Context, electionID int64, number int, excludeCandidateID *int64) (bool, error) {
	for id, c := range m.candidates {
		if c.ElectionID == electionID && c.Number == number {
			if excludeCandidateID != nil && id == *excludeCandidateID {
				continue
			}
			return true, nil
		}
	}
	return false, nil
}

// Mock stats provider
type mockStatsProvider struct{}

func (m *mockStatsProvider) GetCandidateStats(ctx context.Context, electionID int64) (candidate.CandidateStatsMap, error) {
	return candidate.CandidateStatsMap{}, nil
}

func TestAdminCreateCandidate(t *testing.T) {
	repo := newMockRepo()
	stats := &mockStatsProvider{}
	svc := candidate.NewService(repo, stats)

	ctx := context.Background()
	electionID := int64(1)

	req := candidate.AdminCreateCandidateRequest{
		Number:           1,
		Name:             "Test Candidate",
		PhotoURL:         "https://example.com/photo.jpg",
		ShortBio:         "Short bio",
		LongBio:          "Long bio",
		Tagline:          "Test tagline",
		FacultyName:      "Test Faculty",
		StudyProgramName: "Test Program",
		Vision:           "Test vision",
		Missions:         []string{"Mission 1", "Mission 2"},
		MainPrograms:     []candidate.MainProgram{},
		Media:            candidate.Media{},
		SocialLinks:      []candidate.SocialLink{},
		Status:           candidate.CandidateStatusPending,
	}

	dto, err := svc.AdminCreateCandidate(ctx, electionID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dto.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, dto.Name)
	}

	if dto.Number != req.Number {
		t.Errorf("expected number %d, got %d", req.Number, dto.Number)
	}
}

func TestAdminCreateCandidateDuplicateNumber(t *testing.T) {
	repo := newMockRepo()
	stats := &mockStatsProvider{}
	svc := candidate.NewService(repo, stats)

	ctx := context.Background()
	electionID := int64(1)

	// Create first candidate
	req1 := candidate.AdminCreateCandidateRequest{
		Number:           1,
		Name:             "Candidate 1",
		PhotoURL:         "https://example.com/photo1.jpg",
		ShortBio:         "Bio 1",
		LongBio:          "Long bio 1",
		Tagline:          "Tagline 1",
		FacultyName:      "Faculty 1",
		StudyProgramName: "Program 1",
		Vision:           "Vision 1",
		Missions:         []string{"Mission 1"},
		MainPrograms:     []candidate.MainProgram{},
		Media:            candidate.Media{},
		SocialLinks:      []candidate.SocialLink{},
		Status:           candidate.CandidateStatusPending,
	}

	_, err := svc.AdminCreateCandidate(ctx, electionID, req1)
	if err != nil {
		t.Fatalf("first create should succeed, got %v", err)
	}

	// Try to create second candidate with same number
	req2 := req1
	req2.Name = "Candidate 2"

	_, err = svc.AdminCreateCandidate(ctx, electionID, req2)
	if err != candidate.ErrCandidateNumberTaken {
		t.Errorf("expected ErrCandidateNumberTaken, got %v", err)
	}
}

func TestAdminUpdateCandidate(t *testing.T) {
	repo := newMockRepo()
	stats := &mockStatsProvider{}
	svc := candidate.NewService(repo, stats)

	ctx := context.Background()
	electionID := int64(1)

	// Create candidate first
	createReq := candidate.AdminCreateCandidateRequest{
		Number:           1,
		Name:             "Original Name",
		PhotoURL:         "https://example.com/photo.jpg",
		ShortBio:         "Original bio",
		LongBio:          "Original long bio",
		Tagline:          "Original tagline",
		FacultyName:      "Original Faculty",
		StudyProgramName: "Original Program",
		Vision:           "Original vision",
		Missions:         []string{"Mission 1"},
		MainPrograms:     []candidate.MainProgram{},
		Media:            candidate.Media{},
		SocialLinks:      []candidate.SocialLink{},
		Status:           candidate.CandidateStatusPending,
	}

	created, err := svc.AdminCreateCandidate(ctx, electionID, createReq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Update
	newName := "Updated Name"
	updateReq := candidate.AdminUpdateCandidateRequest{
		Name: &newName,
	}

	updated, err := svc.AdminUpdateCandidate(ctx, electionID, created.ID, updateReq)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}
}

func TestAdminDeleteCandidate(t *testing.T) {
	repo := newMockRepo()
	stats := &mockStatsProvider{}
	svc := candidate.NewService(repo, stats)

	ctx := context.Background()
	electionID := int64(1)

	// Create candidate
	createReq := candidate.AdminCreateCandidateRequest{
		Number:           1,
		Name:             "To Be Deleted",
		PhotoURL:         "https://example.com/photo.jpg",
		ShortBio:         "Bio",
		LongBio:          "Long bio",
		Tagline:          "Tagline",
		FacultyName:      "Faculty",
		StudyProgramName: "Program",
		Vision:           "Vision",
		Missions:         []string{"Mission 1"},
		MainPrograms:     []candidate.MainProgram{},
		Media:            candidate.Media{},
		SocialLinks:      []candidate.SocialLink{},
		Status:           candidate.CandidateStatusPending,
	}

	created, err := svc.AdminCreateCandidate(ctx, electionID, createReq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Delete
	err = svc.AdminDeleteCandidate(ctx, electionID, created.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify deleted
	_, err = svc.AdminGetCandidate(ctx, electionID, created.ID)
	if err != candidate.ErrCandidateNotFound {
		t.Errorf("expected ErrCandidateNotFound after delete, got %v", err)
	}
}

func TestAdminPublishUnpublish(t *testing.T) {
	repo := newMockRepo()
	stats := &mockStatsProvider{}
	svc := candidate.NewService(repo, stats)

	ctx := context.Background()
	electionID := int64(1)

	// Create draft candidate
	createReq := candidate.AdminCreateCandidateRequest{
		Number:           1,
		Name:             "Test Candidate",
		PhotoURL:         "https://example.com/photo.jpg",
		ShortBio:         "Bio",
		LongBio:          "Long bio",
		Tagline:          "Tagline",
		FacultyName:      "Faculty",
		StudyProgramName: "Program",
		Vision:           "Vision",
		Missions:         []string{"Mission 1"},
		MainPrograms:     []candidate.MainProgram{},
		Media:            candidate.Media{},
		SocialLinks:      []candidate.SocialLink{},
		Status:           candidate.CandidateStatusPending,
	}

	created, err := svc.AdminCreateCandidate(ctx, electionID, createReq)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Publish
	published, err := svc.AdminPublishCandidate(ctx, electionID, created.ID)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	if published.Status != string(candidate.CandidateStatusApproved) {
		t.Errorf("expected status APPROVED, got %s", published.Status)
	}

	// Unpublish
	unpublished, err := svc.AdminUnpublishCandidate(ctx, electionID, created.ID)
	if err != nil {
		t.Fatalf("unpublish failed: %v", err)
	}

	if unpublished.Status != string(candidate.CandidateStatusPending) {
		t.Errorf("expected status PENDING, got %s", unpublished.Status)
	}
}
