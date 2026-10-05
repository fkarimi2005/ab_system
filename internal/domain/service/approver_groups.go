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

func NewApproverGroupService(g repository.ApproverGroupRepository, u repository.UserRepository) *ApproverGroupService {
	return &ApproverGroupService{groups: g, users: u}
}

func (s *ApproverGroupService) userByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	u, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *ApproverGroupService) Get(
	ctx context.Context, actor models.Actor, experimenterID uuid.UUID,
) (models.ApproverGroup, error) {
	if !actor.IsAdmin() && actor.ID != experimenterID {
		return models.ApproverGroup{}, errs.ErrPermissionDenied
	}
	g, err := s.groups.GetApproverGroup(ctx, experimenterID)
	if err != nil {
		return models.ApproverGroup{}, err
	}
	return g, nil
}

func (s *ApproverGroupService) Set(
	ctx context.Context, experimenterID uuid.UUID, approverIDs []uuid.UUID, minApprovals int,
) (models.ApproverGroup, error) {
	owner, err := s.userByID(ctx, experimenterID)
	if err != nil {
		return models.ApproverGroup{}, err
	}
	if owner.Role.RoleName != models.RoleExperimenter && owner.Role.RoleName != models.RoleAdmin {
		return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
	}
	if len(approverIDs) == 0 || minApprovals < 1 || minApprovals > len(approverIDs) {
		return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
	}

	seen := make(map[uuid.UUID]bool, len(approverIDs))
	g := models.ApproverGroup{ExperimenterID: experimenterID, MinApprovals: minApprovals}
	for _, id := range approverIDs {
		if seen[id] || id == experimenterID {
			return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
		}
		seen[id] = true
		u, err := s.userByID(ctx, id)
		if err != nil {
			return models.ApproverGroup{}, err
		}
		if !u.IsActive || (u.Role.RoleName != models.RoleApprover && u.Role.RoleName != models.RoleAdmin) {
			return models.ApproverGroup{}, errs.ErrInvalidApproverGroup
		}
		g.Members = append(g.Members, models.ApproverGroupMember{ApproverID: id})
	}
	if err := s.groups.SaveApproverGroup(ctx, &g); err != nil {
		return models.ApproverGroup{}, err
	}
	return g, nil
}
