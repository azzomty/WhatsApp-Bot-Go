#!/bin/bash
SESSION=$(curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions" -F "queue_prio=0" -F "partner_id=413" -F "device_id=1869b43f-89bb-48d8-bfaf-0653c9466114" -F "session_prio=0" -F "sensitivity_mode=1" -F "learning_mode=0" -F "origin=US" -F "ft_hard_constraint=ETAT<>'AV'" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json" | grep -oP '"sessionId":"\s*\K[^"]+')
echo "Session: $SESSION"
for i in {1..12}; do
  echo "Step $i"
  RESP=$(curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/answer-question" -F "queue_prio=0" -F "step=$i" -F "answer=0" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json")
  TROUV=$(echo "$RESP" | grep -oP '"trouvitude":\K[^,]+')
  echo "Trouvitude: $TROUV"
  if (( $(echo "$TROUV > 80.0" | bc -l 2>/dev/null || echo 0) )); then
    echo "Guessing..."
    curl -k -s -X POST "https://srv17.akinator.com:9641/game/sessions/$SESSION/set-and-get-toplist" -F "queue_prio=0" -F "step=$i" -F "size=1" -F "add_question_mode=0" -F "trappable_user=0" -H "User-Agent: okhttp/5.4.0" -H "Accept: application/json"
    break
  fi
done
