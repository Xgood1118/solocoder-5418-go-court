#!/bin/bash
BASE=http://localhost:18080
JQ() { node -e "try{const d=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write((d$1)||'')}catch(e){process.stderr.write('parse err: '+e.message+'\n')}"; }

# Fresh courtroom + judge + case for postpone red warn test
echo "=== Setup: fresh courtroom/judge/case for red warn test ==="
CR=$(curl -s -X POST $BASE/courtrooms -H "Content-Type: application/json" -d '{
  "name":"大法庭D","size":"large","capacity":50,
  "equipment":{"projector":true,"recording":true,"interpretation":true,"detention_access":true}
}')
CR_ID=$(echo "$CR" | JQ '.id')
J=$(curl -s -X POST $BASE/judges -H "Content-Type: application/json" -d '{"name":"赵法官","case_types":["criminal"]}')
J_ID=$(echo "$J" | JQ '.id')
C=$(curl -s -X POST $BASE/cases -H "Content-Type: application/json" -d '{
  "case_number":"(2026)刑初999","case_type":"criminal","title":"王五盗窃案",
  "courtroom_size":"large"
}')
C_ID=$(echo "$C" | JQ '.id')
echo "CR_ID=$CR_ID J_ID=$J_ID C_ID=$C_ID"

echo "=== Schedule fresh hearing for 2026-07-10 morning ==="
H=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C_ID\",\"judge_id\":\"$J_ID\",\"courtroom_id\":\"$CR_ID\",
  \"date\":\"2026-07-10\",\"time_slot\":\"morning\",\"duration_min\":60
}")
echo "$H"
H_ID=$(echo "$H" | JQ '.hearing?.id||""')
echo "H_ID=$H_ID"

echo "=== Postpone H to 2026-06-10 (1 day from today 2026-06-09, < 3 days, should RED warn) ==="
curl -s -X POST $BASE/hearings/$H_ID/postpone -H "Content-Type: application/json" -d '{
  "new_date":"2026-06-10","new_time_slot":"morning","reason":"证据未到位"
}'
echo ""

echo "=== DONE ==="