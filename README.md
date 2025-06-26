# GophKeeper

migrate -path=internal/server/storage/migrations -database "postgresql://echo9et:123321@localhost:5432/echo9et?sslmode=disable" -verbose down 1
migrate -path=internal/server/storage/migrations -database "postgresql://echo9et:123321@localhost:5432/echo9et?sslmode=disable" -verbose up
