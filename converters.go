package translatable

import (
	"time"

	"github.com/google/uuid"
)

type TranslatableConverter struct{}

func (c *TranslatableConverter) CreateDTOToModel(dto TranslatableCreateDTO) Translatable {
	translatableID, _ := uuid.Parse(dto.TranslatableID)

	return Translatable{
		ID:             uuid.New(),
		TranslatableID: translatableID,
		Translatable:   dto.Translatable,
		Locale:         dto.Locale,
		Content:        dto.Content,
		CreatedAt:      time.Now(),
	}
}

func (c *TranslatableConverter) UpdateDTOToModel(dto TranslatableUpdateDTO) Translatable {
	return Translatable{
		Locale:  dto.Locale,
		Content: dto.Content,
	}
}

func (c *TranslatableConverter) ModelToResponseDTO(model Translatable) TranslatableResponseDTO {
	return TranslatableResponseDTO(model)
}

func (c *TranslatableConverter) ModelsToResponseDTOs(models []Translatable) []TranslatableResponseDTO {
	dtos := make([]TranslatableResponseDTO, len(models))
	for i, model := range models {
		dtos[i] = c.ModelToResponseDTO(model)
	}
	return dtos
}
