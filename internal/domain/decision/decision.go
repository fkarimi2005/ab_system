package decision

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"

	"AB_system/internal/domain/models"

	"github.com/google/uuid"
)

func Bucket(experimentID uuid.UUID, subjectID string) int {
	sum := sha256.Sum256([]byte(experimentID.String() + ":" + subjectID))
	return int(binary.BigEndian.Uint64(sum[:8]) % models.FullAudienceBP)
}

func Select(variants []models.ExperimentVariant, audienceBP, bucket int) (models.ExperimentVariant, bool) {
	if bucket >= audienceBP {
		return models.ExperimentVariant{}, false
	}
	sorted := append([]models.ExperimentVariant(nil), variants...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	acc := 0
	for _, v := range sorted {
		acc += v.Weight
		if bucket < acc {
			return v, true
		}
	}
	return models.ExperimentVariant{}, false
}

func ID(experimentID uuid.UUID, version int, subjectID, flagKey string) uuid.UUID {
	name := experimentID.String() + ":" + strconv.Itoa(version) + ":" + subjectID + ":" + flagKey
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name))
}
