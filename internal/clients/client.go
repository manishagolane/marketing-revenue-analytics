package clients

import (
	"marketing-revenue-analytics/models"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Clients struct {
	SqsClient *SqsClient
}

func InitializeClients(logger *zap.Logger, queries *models.Queries, conn *pgxpool.Pool) *Clients {
	return &Clients{
		SqsClient: NewSqsClient(logger),
	}
}
