package account

import (
	core "AgriTrace/Internal/Core"
	event_bus "AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"fmt"
	"time"
)

func GetAccountCreateEffect(
	accountID string,
	usersID string,
	name string,
	email string,
	passwordHash string,
	phone string,
	createdAt time.Time,
	updatedAt time.Time,
) []generic.Effect {

	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
                INSERT INTO accounts (account_id, users_id, name, email, password_hash, phone, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            `,
			Args: []any{
				accountID,
				usersID,
				name,
				email,
				passwordHash,
				phone,
				createdAt,
				updatedAt,
			},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Account %s created at %v", accountID, createdAt),
		},
	}
}
func GetAccountUpdateEffect(
	accountID string,
	name string,
	email string,
	phone string,
	updatedAt time.Time,
) []generic.Effect {

	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
                UPDATE accounts
                SET name = $1, email = $2, phone = $3, updated_at = $4
                WHERE account_id = $5
            `,
			Args: []any{
				name,
				email,
				phone,
				updatedAt,
				accountID,
			},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Account %s updated at %v", accountID, updatedAt),
		},
	}
}

func ListenAccount(b *event_bus.EventBus, topic, workerTopic string, jobStore *generic.JobStore) {
    sub := b.Subscribe(topic)

    go func(jobStore *generic.JobStore) {
        for event := range sub {

            var effects []generic.Effect

            switch event.SubTopic {

            case core.AccountCreated:
                payload, ok := event.Payload.(core.AccountCreatedReq)
                if !ok {
                    fmt.Println("Invalid payload for AccountCreated:", event.Payload)
                    continue
                }

                effects = GetAccountCreateEffect(
                    payload.AccountID,
                    payload.UsersID,
                    payload.Name,
                    payload.Email,
                    payload.PasswordHash,
                    payload.Phone,
                    time.Now().UTC(),
                    time.Now().UTC(),
                )

            case core.AccountUpdated:
                payload, ok := event.Payload.(core.AccountUpdatedReq)
                if !ok {
                    fmt.Println("Invalid payload for AccountUpdated:", event.Payload)
                    continue
                }

                effects = GetAccountUpdateEffect(
                    payload.AccountID,
                    payload.Name,
                    payload.Email,
                    payload.Phone,
                    time.Now().UTC(),
                )
            }

            if effects == nil {
                continue
            }

            // publish ke worker topic
            workEvent := event_bus.Event{
                WorkId:  event.WorkId,
                Payload: effects,
            }

            b.Publish(workerTopic, workEvent)
        }
    }(jobStore)
}
