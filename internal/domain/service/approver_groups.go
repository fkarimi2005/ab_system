package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/repository"
	"AB_system/pkg/errs"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ApproverGroupService struct {
	groups repository.ApproverGroupRepository
	users  repository.UserRepository
}

func NewApproverGroupService(
	groups repository.ApproverGroupRepository,
	users repository.UserRepository,
) *ApproverGroupService {
	return &ApproverGroupService{groups: groups, users: users}
}

func (s *ApproverGroupService) user(ctx context.Context, id uuid.UUID) (models.User, error) {
	u, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

// Get: смотреть группу может admin или сам экспериментатор.
func (s *ApproverGroupService) Get(
	ctx context.Context, actor models.Actor, experimenterID uuid.UUID,
) (models.ApproverGroup, error) {
	if !actor.IsAdmin() && actor.ID != experimenterID {
		return models.ApproverGroup{}, errs.ErrPermissionDenied
	}
	return s.groups.GetApproverGroup(ctx, experimenterID)
}

func (s *ApproverGroupService) Set(
	ctx context.Context, experimenterID uuid.UUID, approverIDs []uuid.UUID, minApprovals int,
) (models.ApproverGroup, error) {
	owner, err := s.user(ctx, experimenterID)
	if err != nil {
		return models.ApproverGroup{}, err
	}
	if owner.Role.RoleName != models.RoleExperimenter && owner.Role.RoleName != models.RoleAdmin {
		return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
	}
	if len(approverIDs) == 0 || minApprovals < 1 || minApprovals > len(approverIDs) {
		return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
	}

	g := models.ApproverGroup{ExperimenterID: experimenterID, MinApprovals: minApprovals}
	seen := make(map[uuid.UUID]bool, len(approverIDs))
	for _, id := range approverIDs {
		if seen[id] || id == experimenterID { // дубли и сам владелец не допускаются
			return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
		}
		seen[id] = true

		u, err := s.user(ctx, id)
		if err != nil {
			return models.ApproverGroup{}, err
		}
		isApprover := u.Role.RoleName == models.RoleApprover || u.Role.RoleName == models.RoleAdmin
		if !u.IsActive || !isApprover {
			return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
		}
		g.Members = append(g.Members, models.ApproverGroupMember{ApproverID: id})
	}

	if err := s.groups.SaveApproverGroup(ctx, &g); err != nil {
		return models.ApproverGroup{}, err
	}
	return g, nil
}
