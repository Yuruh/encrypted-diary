set -e

#docker-compose -f docker-compose.test.yml build

docker-compose -f docker-compose.test.yml -p diary-tests up -d

#replace ./src by whatever you want to test
docker exec -it diary-tests_diary-api_1 go test ./... -cover -coverprofile=coverage.out

docker exec -it diary-tests_diary-api_1 go tool cover -html=coverage.out -o coverage.html

#docker-compose -f docker-compose.test.yml -p diary-tests down
