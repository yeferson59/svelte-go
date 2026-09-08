package market

import (
	"context"
	"errors"
	"time"

	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/yeferson59/gofinance/v2/decimal"
	"github.com/yeferson59/gofinance/v2/money"
)

// credentialCols is the safe projection: every column except the sealed
// material. Queries that serve a handler use this one.
const credentialCols = `provider, last4, status, last_verified_at, COALESCE(last_error, ''), created_at, updated_at`

func scanCredential(row pgx.Row) (Credential, error) {
	var c Credential
	if err := row.Scan(&c.Provider, &c.Last4, &c.Status, &c.LastVerifiedAt, &c.LastError, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return Credential{}, err
	}

	return c, nil
}

// UpsertCredential stores a sealed key, replacing whatever the user had for
// that provider. The previous error is always cleared: the new key has not been
// proven bad yet.
//
// status comes from the caller because saving a key proves something about it.
// It is normally 'active'; a key saved while its quota is exhausted is stored
// 'rate_limited', which is the honest state and is what the UI shows.
func (r *PostgresRepository) UpsertCredential(ctx context.Context, userID uuid.UUID, cred sealedCredential, keyLast4 string, status CredentialStatus, verifiedAt *time.Time) (Credential, error) {
	return scanCredential(r.db.QueryRow(ctx, `
		INSERT INTO market_credentials
			(user_id, provider, kek_version, wrapped_dek, nonce, ciphertext, last4, status, last_verified_at, last_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL)
		ON CONFLICT (user_id, provider) DO UPDATE SET
			kek_version      = EXCLUDED.kek_version,
			wrapped_dek      = EXCLUDED.wrapped_dek,
			nonce            = EXCLUDED.nonce,
			ciphertext       = EXCLUDED.ciphertext,
			last4            = EXCLUDED.last4,
			status           = EXCLUDED.status,
			last_verified_at = EXCLUDED.last_verified_at,
			last_error       = NULL,
			updated_at       = NOW()
		RETURNING `+credentialCols,
		userID, string(cred.Provider), cred.Sealed.KEKVersion,
		cred.Sealed.WrappedDEK, cred.Sealed.Nonce, cred.Sealed.Ciphertext,
		keyLast4, string(status), verifiedAt,
	))
}

// ListCredentials returns what the user may see about their own keys. The
// sealed columns are not selected at all, so there is no path from this query
// to a leaked ciphertext.
func (r *PostgresRepository) ListCredentials(ctx context.Context, userID uuid.UUID) ([]Credential, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+credentialCols+`
		FROM market_credentials WHERE user_id = $1 ORDER BY provider
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	creds := make([]Credential, 0)
	for rows.Next() {
		c, err := scanCredential(rows)
		if err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}

	return creds, rows.Err()
}

// GetSealedCredentials loads the encrypted keys a user has, for the service to
// open just long enough to build a provider chain. Only the sync path calls it.
func (r *PostgresRepository) GetSealedCredentials(ctx context.Context, userID uuid.UUID) ([]sealedCredential, error) {
	rows, err := r.db.Query(ctx, `
		SELECT provider, kek_version, wrapped_dek, nonce, ciphertext
		FROM market_credentials
		WHERE user_id = $1 AND status <> 'invalid'
		ORDER BY provider
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	creds := make([]sealedCredential, 0)
	for rows.Next() {
		var c sealedCredential
		if err := rows.Scan(&c.Provider, &c.Sealed.KEKVersion, &c.Sealed.WrappedDEK, &c.Sealed.Nonce, &c.Sealed.Ciphertext); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}

	return creds, rows.Err()
}

// GetSealedCredential loads one provider's key, used by the verify endpoint.
func (r *PostgresRepository) GetSealedCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) (sealedCredential, error) {
	var c sealedCredential

	err := r.db.QueryRow(ctx, `
		SELECT provider, kek_version, wrapped_dek, nonce, ciphertext
		FROM market_credentials WHERE user_id = $1 AND provider = $2
	`, userID, string(provider)).Scan(&c.Provider, &c.Sealed.KEKVersion, &c.Sealed.WrappedDEK, &c.Sealed.Nonce, &c.Sealed.Ciphertext)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sealedCredential{}, ErrCredentialNotFound
		}

		return sealedCredential{}, err
	}

	return c, nil
}

func (r *PostgresRepository) DeleteCredential(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM market_credentials WHERE user_id = $1 AND provider = $2
	`, userID, string(provider))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCredentialNotFound
	}

	return nil
}

// MarkCredentialWorking retires a stale verdict after a call that actually
// succeeded with this key.
//
// The WHERE clause carries the whole design. It makes the common case — a key
// that was already active — a no-op the database resolves on the index, so this
// can run after every successful sync without becoming a write per price. And
// unlike SetCredentialStatus a missing row is not an error: the caller learns
// the provider from the answer it just got, so being asked about a credential
// that is no longer there means the user deleted it mid-run, not that anything
// is wrong.
func (r *PostgresRepository) MarkCredentialWorking(ctx context.Context, userID uuid.UUID, provider ProviderID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE market_credentials
		SET status = 'active',
		    last_error = NULL,
		    last_verified_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = $1 AND provider = $2 AND status <> 'active'
	`, userID, string(provider))

	return err
}

// SetCredentialStatus records the verdict of a provider call. lastErr is
// expected to be already scrubbed of the key by the provider client.
//
// A verdict about a credential the user does not have is a bug in the caller,
// not a no-op, so a missing row is reported rather than swallowed.
func (r *PostgresRepository) SetCredentialStatus(ctx context.Context, userID uuid.UUID, provider ProviderID, status CredentialStatus, lastErr string) error {
	var verifiedAt *time.Time
	if status == CredentialActive {
		verifiedAt = new(time.Now().UTC())
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE market_credentials
		SET status = $3,
		    last_error = NULLIF($4, ''),
		    last_verified_at = COALESCE($5, last_verified_at),
		    updated_at = NOW()
		WHERE user_id = $1 AND provider = $2
	`, userID, string(provider), string(status), lastErr, verifiedAt)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrCredentialNotFound
	}

	return nil
}

// UsersWithCredentials lists the users the sync job should walk. Users whose
// only keys are known-invalid are skipped: retrying them every morning just
// burns requests against a key that is already known not to work.
func (r *PostgresRepository) UsersWithCredentials(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT user_id FROM market_credentials WHERE status <> 'invalid'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// UpsertUserAssetPrice stores a price against the user whose key fetched it.
// The currency comes from the asset rather than from the Money value, matching
// how UpdateAssetPrice already writes the shared column.
func (r *PostgresRepository) UpsertUserAssetPrice(ctx context.Context, userID, assetID uuid.UUID, price money.Money, currency money.Currency, source ProviderID, fetchedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_asset_prices (user_id, asset_id, price, currency, source, fetched_at)
		VALUES ($1, $2, $3::numeric, $4, $5, $6)
		ON CONFLICT (user_id, asset_id) DO UPDATE SET
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			source = EXCLUDED.source,
			fetched_at = EXCLUDED.fetched_at
	`, userID, assetID, price.String(), currency, string(source), fetchedAt)

	return err
}

// UpsertUserExchangeRate stores a rate against the user whose key fetched it.
func (r *PostgresRepository) UpsertUserExchangeRate(ctx context.Context, userID uuid.UUID, from, to money.Currency, rate decimal.Decimal, source ProviderID, fetchedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_exchange_rates (user_id, from_currency, to_currency, rate, source, fetched_at)
		VALUES ($1, $2, $3, $4::numeric, $5, $6)
		ON CONFLICT (user_id, from_currency, to_currency) DO UPDATE SET
			rate = EXCLUDED.rate,
			source = EXCLUDED.source,
			fetched_at = EXCLUDED.fetched_at
	`, userID, from, to, rate.String(), string(source), fetchedAt)

	return err
}
