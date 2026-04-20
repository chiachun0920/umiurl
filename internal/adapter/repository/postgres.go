package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"umiurl/internal/domain"
	repoiface "umiurl/internal/usecase/interface/repository"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, shortURL domain.ShortURL) (domain.ShortURL, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO short_urls (
			code, original_url, referral_code, campaign,
			preview_title, preview_description, preview_image_url,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`,
		shortURL.Code,
		shortURL.OriginalURL,
		shortURL.ReferralCode,
		shortURL.Campaign,
		shortURL.Preview.Title,
		shortURL.Preview.Description,
		shortURL.Preview.ImageURL,
		shortURL.CreatedAt,
		shortURL.UpdatedAt,
	).Scan(&shortURL.ID, &shortURL.CreatedAt, &shortURL.UpdatedAt)
	if isUniqueViolation(err) {
		return domain.ShortURL{}, repoiface.ErrDuplicateCode
	}
	if err != nil {
		return domain.ShortURL{}, err
	}
	return shortURL, nil
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (domain.ShortURL, bool, error) {
	var entity domain.ShortURL
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, original_url, referral_code, campaign,
			preview_title, preview_description, preview_image_url,
			created_at, updated_at
		FROM short_urls
		WHERE code = $1
	`, code).Scan(
		&entity.ID,
		&entity.Code,
		&entity.OriginalURL,
		&entity.ReferralCode,
		&entity.Campaign,
		&entity.Preview.Title,
		&entity.Preview.Description,
		&entity.Preview.ImageURL,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShortURL{}, false, nil
	}
	if err != nil {
		return domain.ShortURL{}, false, err
	}
	return entity, true, nil
}

func (r *PostgresRepository) UpdatePreview(ctx context.Context, id int64, preview domain.PreviewMetadata, updatedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE short_urls
		SET preview_title = $1,
			preview_description = $2,
			preview_image_url = $3,
			updated_at = $4
		WHERE id = $5
	`,
		preview.Title,
		preview.Description,
		preview.ImageURL,
		updatedAt,
		id,
	)
	return err
}

func (r *PostgresRepository) Record(ctx context.Context, event domain.ClickEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO click_events (
			short_url_id, referral_code, campaign, user_agent,
			platform, device, country, occurred_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		event.ShortURLID,
		event.ReferralCode,
		event.Campaign,
		event.UserAgent,
		event.Platform,
		event.Device,
		event.Country,
		event.OccurredAt,
	)
	return err
}

func (r *PostgresRepository) RecordConversion(ctx context.Context, event domain.ConversionEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO conversion_events (
			short_url_id, referral_code, campaign, value, occurred_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		event.ShortURLID,
		event.ReferralCode,
		event.Campaign,
		event.Value,
		event.OccurredAt,
	)
	return err
}

func (r *PostgresRepository) Summary(ctx context.Context, code string) (domain.AnalyticsSummary, bool, error) {
	var summary domain.AnalyticsSummary
	var shortURLID int64
	err := r.pool.QueryRow(ctx, `SELECT id, code FROM short_urls WHERE code = $1`, code).Scan(&shortURLID, &summary.Code)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AnalyticsSummary{}, false, nil
	}
	if err != nil {
		return domain.AnalyticsSummary{}, false, err
	}

	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM click_events WHERE short_url_id = $1`, shortURLID).Scan(&summary.TotalClicks); err != nil {
		return domain.AnalyticsSummary{}, false, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM conversion_events WHERE short_url_id = $1`, shortURLID).Scan(&summary.TotalConversions); err != nil {
		return domain.AnalyticsSummary{}, false, err
	}

	var errBreakdowns error
	summary.ClicksByReferral, errBreakdowns = r.breakdown(ctx, "click_events", "referral_code", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ClicksByCampaign, errBreakdowns = r.breakdown(ctx, "click_events", "campaign", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ClicksByPlatform, errBreakdowns = r.breakdown(ctx, "click_events", "platform", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ClicksByDevice, errBreakdowns = r.breakdown(ctx, "click_events", "device", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ClicksByCountry, errBreakdowns = r.breakdown(ctx, "click_events", "country", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ConversionsByRef, errBreakdowns = r.breakdown(ctx, "conversion_events", "referral_code", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}
	summary.ConversionsByCamp, errBreakdowns = r.breakdown(ctx, "conversion_events", "campaign", shortURLID)
	if errBreakdowns != nil {
		return domain.AnalyticsSummary{}, false, errBreakdowns
	}

	return summary, true, nil
}

func (r *PostgresRepository) breakdown(ctx context.Context, table, column string, shortURLID int64) ([]domain.Breakdown, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT COALESCE(NULLIF(`+column+`, ''), 'unknown') AS key, count(*) AS count
		FROM `+table+`
		WHERE short_url_id = $1
		GROUP BY key
		ORDER BY count DESC, key ASC
	`, shortURLID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Breakdown
	for rows.Next() {
		var item domain.Breakdown
		if err := rows.Scan(&item.Key, &item.Count); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
