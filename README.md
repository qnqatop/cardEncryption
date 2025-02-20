## Перед запуском

1 - поднять контейнер

2 - зайти в образ vault и прописать  `docker-compose exec vault vault login myroot`

3 - выполнить команду `docker-compose exec vault vault secrets enable -path=transit transit` (создает путь до dek)
