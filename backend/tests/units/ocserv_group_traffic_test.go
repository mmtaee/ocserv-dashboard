package units

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/database"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/repository"
	"github.com/mmtaee/ocserv-dashboard/backend/pkg/request"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOcservGroupsIncludeAggregatedUserTraffic(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	originalDB := database.PostgresDB
	database.PostgresDB = db
	t.Cleanup(func() { database.PostgresDB = originalDB })

	mock.ExpectQuery(`SELECT count\(\*\) FROM "ocserv_groups"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`(?s)SELECT .*COALESCE\(user_traffic\.total_rx, 0\) AS total_rx.*COALESCE\(user_traffic\.total_tx, 0\) AS total_tx.*LEFT JOIN \(SELECT "group", SUM\(running_rx\) AS total_rx, SUM\(running_tx\) AS total_tx FROM "ocserv_users" GROUP BY "group"\) AS user_traffic ON user_traffic\."group" = ocserv_groups\.name.*`).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "config", "total_rx", "total_tx"}).
			AddRow(1, "multiple-users", nil, int64(120+80), int64(250+50)).
			AddRow(2, "single-user", nil, int64(75), int64(125)).
			AddRow(3, "no-users", nil, int64(0), int64(0)))

	groups, total, err := repository.NewOcservGroupRepository().Groups(context.Background(), &request.Pagination{
		Page: 1, PageSize: 10, Order: "id", Sort: "ASC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, groups, 3)
	require.Equal(t, int64(200), groups[0].TotalRX)
	require.Equal(t, int64(300), groups[0].TotalTX)
	require.Equal(t, int64(75), groups[1].TotalRX)
	require.Equal(t, int64(125), groups[1].TotalTX)
	require.Zero(t, groups[2].TotalRX)
	require.Zero(t, groups[2].TotalTX)
	mock.ExpectClose()
	require.NoError(t, sqlDB.Close())
	require.NoError(t, mock.ExpectationsWereMet())
}
