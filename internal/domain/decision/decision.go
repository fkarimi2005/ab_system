// Package decision — чистая логика выбора варианта эксперимента для пользователя.
// Без базы и без времени: одинаковые входные данные всегда дают одинаковый результат.
package decision

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"

	"AB_system/internal/domain/models"

	"github.com/google/uuid"
)

// Bucket относит пользователя к корзине 0..9999.
// Соль — id эксперимента: в разных экспериментах один и тот же пользователь
// попадает в разные корзины, поэтому эксперименты не коррелируют между собой.
func Bucket(experimentID uuid.UUID, subjectID string) int {
	sum := sha256.Sum256([]byte(experimentID.String() + ":" + subjectID))
	return int(binary.BigEndian.Uint64(sum[:8]) % models.FullAudienceBP)
}

// Select выбирает вариант по корзине. ok == false — пользователь вне аудитории.
//
// Корзины [0, audienceBP) делятся между вариантами по весам в порядке имён.
// Сумма весов равна audienceBP (это проверяется при сохранении эксперимента),
// поэтому каждая корзина внутри аудитории попадает ровно в один вариант.
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
	return models.ExperimentVariant{}, false // веса не покрывают аудиторию: считаем «вне»
}

// ID — детерминированный идентификатор решения (UUIDv5): тот же запрос
// при той же версии эксперимента даёт тот же decision_id.
func ID(experimentID uuid.UUID, version int, subjectID, flagKey string) uuid.UUID {
	name := experimentID.String() + ":" + strconv.Itoa(version) + ":" + subjectID + ":" + flagKey
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name))
}
