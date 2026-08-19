#!/bin/bash
SESSION=$(curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions" -F "queue_prio=0" -F "partner_id=413" -F "device_id=1869b43f-89bb-48d8-bfaf-0653c9466114" -F "session_prio=0" -F "sensitivity_mode=1" -F "learning_mode=0" -F "origin=US" -F "ft_hard_constraint=ETAT<>'AV'" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json" | grep -oP '"sessionId":"\s*\K[^"]+')
for i in {1..7}; do
  curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/answer-question" -F "queue_prio=0" -F "step=$i" -F "answer=0" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json" > /dev/null
done
curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/set-and-get-toplist" -F "queue_prio=0" -F "step=8" -F "size=1" -F "add_question_mode=0" -F "trappable_user=0" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json" > /dev/null
curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/decline-toplist" -F "queue_prio=0" -F "step=9" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json" > /dev/null
echo "Testing next step API..."
curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/answer-question" -F "queue_prio=0" -F "step=10" -F "answer=0" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json"
