# completed by SimpleMouse / zhindasp

# What was the task?
In the previous project (https://github.com/S1mpleM0use/s21_tictactoe_with_b64_auth), the goal was to add Base64 auth, add a persistent storage (PostgreSQL) with User support and Multiplayer capabilities. 
The task in AP01_T05B was to:
1. To switch out the Base64 auth with JWT token-based one
2. To add a Leaderboard
3. To add Game History

# RUN A GAME:
cd src
docker compose up --build

# STOP:
docker compose down

# STOP AND DELETE:
docker compose down -v

# TEST A GAME:
cd src
go test -v -race ./...

# User endpoints

## register user:
curl -X POST http://localhost:8080/user -H 'Content-Type: application/json' -d '{"login": "test_user_1", "password": "121212"}'

## login:
curl -X POST http://localhost:8080/user/login \
  -H 'Content-Type: application/json' \
  -d '{"login": "test_user_1", "password": "121212"}'

# SAVE ACCESS TOKEN AS A VARIABLE:
TOKEN=$(curl -s -X POST http://localhost:8080/user/login \
  -H 'Content-Type: application/json' \
  -d '{"login": "test_user_1", "password": "121212"}' \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

## refresh access token:
curl -X POST http://localhost:8080/user/refresh-access \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token": "REFRESH_TOKEN_HERE"}'

## refresh refresh token (requires valid access token):
curl -X POST http://localhost:8080/user/refresh-token \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token": "REFRESH_TOKEN_HERE"}'

## get info about yourself:
curl http://localhost:8080/user/me \
  -H "Authorization: Bearer $TOKEN"

## get user info by ID:
curl http://localhost:8080/user/PLAYER_ID \
  -H "Authorization: Bearer $TOKEN"

# Game methods

## create a game vs player:
curl -X POST http://localhost:8080/game \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"mode": "VSplayer"}'

## vs computer:
curl -X POST http://localhost:8080/game \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"mode": "VScomputer"}'

## list games available to join:
curl http://localhost:8080/game/list \
  -H "Authorization: Bearer $TOKEN"

## join a game:
curl -X POST http://localhost:8080/game/GAME_ID/join \
  -H "Authorization: Bearer $TOKEN_2"

## make a move:
curl -X POST http://localhost:8080/game/GAME_ID/move \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"row": 0, "column": 0}'

## get a current game:
curl http://localhost:8080/game/current \
  -H "Authorization: Bearer $TOKEN"

## finished user games:
curl http://localhost:8080/game/finished \
  -H "Authorization: Bearer $TOKEN"

# Leaderboard

## top 10:
curl http://localhost:8080/game/leaderboard \
  -H "Authorization: Bearer $TOKEN"

## top 5 using query:
curl "http://localhost:8080/game/leaderboard?limit=5" \
  -H "Authorization: Bearer $TOKEN"

## top 3:
curl http://localhost:8080/game/leaderboard/3 \
  -H "Authorization: Bearer $TOKEN"
