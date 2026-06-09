#!/bin/bash
BASE=http://localhost:18080
JQ() { node -e "try{const d=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write((d$1)||'')}catch(e){process.stderr.write('parse err: '+e.message+'\n')}"; }

echo "=== 1. Create criminal courtroom (large, detention_access=true) ==="
CR1=$(curl -s -X POST $BASE/courtrooms -H "Content-Type: application/json" -d '{
  "name":"大法庭A","size":"large","capacity":50,
  "equipment":{"projector":true,"recording":true,"interpretation":true,"detention_access":true}
}')
echo "$CR1"
CR1_ID=$(echo "$CR1" | JQ '.id')
echo "CR1_ID=$CR1_ID"

echo "=== 2. Create civil courtroom (medium, detention_access=false) ==="
CR2=$(curl -s -X POST $BASE/courtrooms -H "Content-Type: application/json" -d '{
  "name":"中法庭B","size":"medium","capacity":20,
  "equipment":{"projector":true,"recording":true,"interpretation":false,"detention_access":false}
}')
echo "$CR2"
CR2_ID=$(echo "$CR2" | JQ '.id')
echo "CR2_ID=$CR2_ID"

echo "=== 3. Create judge (criminal type, max 3/day) ==="
J1=$(curl -s -X POST $BASE/judges -H "Content-Type: application/json" -d '{"name":"张法官","case_types":["criminal"]}')
echo "$J1"
J1_ID=$(echo "$J1" | JQ '.id')
echo "J1_ID=$J1_ID"

echo "=== 4. Create 3 jurors ==="
JU1=$(curl -s -X POST $BASE/jurors -H "Content-Type: application/json" -d '{"name":"陪审员1","case_types":["criminal"]}')
JU1_ID=$(echo "$JU1" | JQ '.id')
JU2=$(curl -s -X POST $BASE/jurors -H "Content-Type: application/json" -d '{"name":"陪审员2","case_types":["criminal"]}')
JU2_ID=$(echo "$JU2" | JQ '.id')
JU3=$(curl -s -X POST $BASE/jurors -H "Content-Type: application/json" -d '{"name":"陪审员3","case_types":["criminal"]}')
JU3_ID=$(echo "$JU3" | JQ '.id')
echo "JU1=$JU1_ID JU2=$JU2_ID JU3=$JU3_ID"

echo "=== 5. Create criminal case with lawyer + witnesses ==="
C1=$(curl -s -X POST $BASE/cases -H "Content-Type: application/json" -d '{
  "case_number":"(2026)刑初001","case_type":"criminal","title":"张三故意伤害案",
  "parties":[{"id":"p1","name":"张三","type":"defendant","phone":"13800000001","email":"a@b.com"}],
  "lawyers":[{"id":"law1","name":"李律师","phone":"13900000001","email":"l@b.com","firm":"正义所"}],
  "witnesses":[{"id":"w1","name":"证人A","phone":"13700000001","email":"w@b.com","witness_type":"defendant_witness"},
               {"id":"w2","name":"鉴定人B","phone":"13600000001","email":"e@b.com","witness_type":"expert"}],
  "courtroom_size":"large"
}')
echo "$C1"
C1_ID=$(echo "$C1" | JQ '.id')
echo "C1_ID=$C1_ID"

echo "=== 6. Create civil case ==="
C2=$(curl -s -X POST $BASE/cases -H "Content-Type: application/json" -d '{
  "case_number":"(2026)民初001","case_type":"civil","title":"李四借款案",
  "lawyers":[{"id":"law2","name":"王律师","phone":"13900000002","email":"w@b.com","firm":"天平所"}],
  "courtroom_size":"medium"
}')
echo "$C2"
C2_ID=$(echo "$C2" | JQ '.id')
echo "C2_ID=$C2_ID"

echo "=== 7. Schedule criminal hearing morning 2026-07-15 (juror_count=2) ==="
H1=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C1_ID\",\"judge_id\":\"$J1_ID\",\"courtroom_id\":\"$CR1_ID\",
  \"date\":\"2026-07-15\",\"time_slot\":\"morning\",\"duration_min\":150,
  \"juror_count\":2,\"translator\":\"翻译甲\",\"expert\":\"鉴定人乙\"
}")
echo "$H1"
H1_ID=$(echo "$H1" | JQ '.hearing?.id||""')
echo "H1_ID=$H1_ID"

echo "=== 8. Try same judge same slot — should conflict ==="
H_DUP=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C1_ID\",\"judge_id\":\"$J1_ID\",\"courtroom_id\":\"$CR2_ID\",
  \"date\":\"2026-07-15\",\"time_slot\":\"morning\",\"duration_min\":60
}")
echo "$H_DUP"

echo "=== 9. Schedule civil case afternoon ==="
H2=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C2_ID\",\"judge_id\":\"$J1_ID\",\"courtroom_id\":\"$CR2_ID\",
  \"date\":\"2026-07-15\",\"time_slot\":\"afternoon\",\"duration_min\":120
}")
echo "$H2"
H2_ID=$(echo "$H2" | JQ '.hearing?.id||""')
echo "H2_ID=$H2_ID"

echo "=== 10. Criminal case in non-detention courtroom — should reject ==="
H3=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C1_ID\",\"judge_id\":\"$J1_ID\",\"courtroom_id\":\"$CR2_ID\",
  \"date\":\"2026-07-16\",\"time_slot\":\"morning\",\"duration_min\":60
}")
echo "$H3"

echo "=== 11. Notices for H1 ==="
sleep 1
curl -s "$BASE/notices?hearing_id=$H1_ID" | head -c 2000
echo ""

echo "=== 12. Notice types ==="
curl -s $BASE/notices/types
echo ""

echo "=== 13. Postpone H1 to 2026-07-20 afternoon (>3 days, no warn) ==="
P1=$(curl -s -X POST $BASE/hearings/$H1_ID/postpone -H "Content-Type: application/json" -d '{
  "new_date":"2026-07-20","new_time_slot":"afternoon","reason":"法官出差"
}')
echo "$P1"

echo "=== 14. Postpone H1 to 2026-07-16 morning (1 day, <3 days = red warn) ==="
P2=$(curl -s -X POST $BASE/hearings/$H1_ID/postpone -H "Content-Type: application/json" -d '{
  "new_date":"2026-07-16","new_time_slot":"morning","reason":"证据未到位"
}')
echo "$P2"

echo "=== 15. Cancel H2 ==="
curl -s -X POST $BASE/hearings/$H2_ID/cancel -H "Content-Type: application/json" -d '{"reason":"和解"}'
echo ""

echo "=== 16. Complete H1 ==="
curl -s -X POST $BASE/hearings/$H1_ID/complete
echo ""

echo "=== 17. List all hearings ==="
curl -s $BASE/hearings | head -c 3000
echo ""

echo "=== 18. Generate monthly stats ==="
curl -s -X POST $BASE/stats/generate -H "Content-Type: application/json" -d '{"month":"2026-07"}' | head -c 2500
echo ""

echo "=== 19. Get stats for 2026-07 ==="
curl -s $BASE/stats/2026-07 | head -c 2500
echo ""

echo "=== 20. Judge daily count ==="
curl -s "$BASE/judges/$J1_ID/daily-count?date=2026-07-15"
echo ""

echo "=== 21. List snapshots ==="
ls -la $BASE 2>/dev/null
echo "=== DONE ==="