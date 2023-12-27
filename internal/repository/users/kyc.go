package users

import (
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/users"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

func (z *usersRepository) GetKYC(paginationParams *types.PaginationParams, searchParams *models.UserKYCSearch) (*database.PaginatedResult, error) {
	var data []models.UserKYCSearch
	var count int64

	db := z.database.Debug().Table(z.repoConfig.KycTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.KycTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.KycTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	for i := range data {
		if data[i].ID == nil {
			continue
		}

		kycID := *data[i].ID
		var kycFiles []models.UserKYCFile
		if err := z.database.Table(z.repoConfig.UserFileTable).Where("kyc_id = ?", kycID).Find(&kycFiles).Error; err != nil {
			z.logger.Error("error loading KYC files", zap.Int64("kyc_id", kycID), zap.Error(err))
			continue
		}

		var filesData []models.FilesOfKYCs
		for _, kycFile := range kycFiles {
			var file models.FilesOfKYCs
			if err := z.database.Table(z.repoConfig.FileServiceTable).Where("id = ?", kycFile.FileID).Find(&file).Error; err != nil {
				z.logger.Error("error loading file service files", zap.Int64("file_id", kycFile.FileID), zap.Error(err))
				continue
			}
			filesData = append(filesData, file)
		}

		data[i].UserFiles = &filesData

		data[i].KYCFiles = &kycFiles
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}
