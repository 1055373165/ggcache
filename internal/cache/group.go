package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	pb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/pkg/common/logger"

	"gorm.io/gorm"
)

// NewGroupManager creates and initializes cache groups for the given group names.
// It returns a map of group names to their respective Group instances.
func NewGroupManager(groupNames []string, currentPeerAddr string) map[string]*Group {
	for _, name := range groupNames {
		retriever := createStudentRetriever()
		group := NewGroup(name, config.Conf.GroupManager.Strategy, config.Conf.GroupManager.MaxCacheSize, retriever)
		GroupManager[name] = group
		logger.LogrusObj.Infof("Group %s created with strategy %s", name, config.Conf.GroupManager.Strategy)
	}

	return GroupManager
}

// createStudentRetriever creates a new RetrieveFunc that fetches student data from the database.
// It includes proper error handling and logging.
func createStudentRetriever() RetrieveFunc {
	return func(key string) ([]byte, error) {
		start := time.Now()
		defer func() {
			logger.LogrusObj.Debugf("Database query time: %v ms", time.Since(start).Milliseconds())
		}()

		ctx := context.Background()
		studentDAO := dao.NewStudentDao(ctx)

		student, err := studentDAO.ShowStudentInfo(&pb.StudentRequest{
			Name: key,
		})

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.LogrusObj.Infof("Student not found in database: %s", key)
				// Return empty bytes for cache negative results
				return []byte{}, nil
			}
			return nil, fmt.Errorf("database query failed: %w", err)
		}

		logger.LogrusObj.Infof("Successfully retrieved student %s score: %.2f", key, student.Score)

		// Format score with 2 decimal places
		score := strconv.FormatFloat(float64(student.Score), 'f', 2, 64)
		return []byte(score), nil
	}
}
