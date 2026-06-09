#!/bin/bash
BASE=http://localhost:18080
JQ() { node -e "try{const d=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write((d$1)||'')}catch(e){process.stderr.write('parse err: '+e.message+'\n')}"; }

# Test conflict: schedule 2nd hearing for SAME judge (same date, same slot) but DIFFERENT courtroom with detention_access
# We need a criminal courtroom with detention_access
echo "=== A. Create 2nd detention courtroom for conflict test ==="
CR3=$(curl -s -X POST $BASE/courtrooms -H "Content-Type: application/json" -d '{
  "name":"大法庭C","size":"large","capacity":50,
  "equipment":{"projector":true,"recording":true,"interpretation":true,"detention_access":true}
}')
echo "$CR3"
CR3_ID=$(echo "$CR3" | JQ '.id')

# need existing judge + case from earlier — read /judges to find any
J_LIST=$(curl -s $BASE/judges)
J_ID=$(echo "$J_LIST" | node -e "const a=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write(a[0]?.id||'')")
echo "J_ID=$J_ID"

C_LIST=$(curl -s $BASE/cases)
C_ID=$(echo "$C_LIST" | node -e "const a=JSON.parse(require('fs').readFileSync(0,'utf8'));process.stdout.write(a.find(c=>c.case_type==='criminal')?.id||'')")
echo "C_ID=$C_ID"

echo "=== B. Schedule same judge same slot 2026-07-15 morning (different courtroom, both detention) — should judge-conflict ==="
curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C_ID\",\"judge_id\":\"$J_ID\",\"courtroom_id\":\"$CR3_ID\",
  \"date\":\"2026-07-15\",\"time_slot\":\"morning\",\"duration_min\":60
}"
echo ""

echo "=== C. Test postpone red warning: schedule hearing then postpone to tomorrow (1 day < 3 days) ==="
# Schedule fresh hearing for 2026-07-10 (1 day from today)
H_NEW=$(curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$C_ID\",\"judge_id\":\"$J_ID\",\"courtroom_id\":\"$CR3_ID\",
  \"date\":\"2026-07-10\",\"time_slot\":\"morning\",\"duration_min\":60
}")
H_NEW_ID=$(echo "$H_NEW" | JQ '.hearing?.id||""')
echo "H_NEW_ID=$H_NEW_ID"
echo "Postpone H_NEW to 2026-06-10 (1 day from today, should red warn)"
curl -s -X POST $BASE/hearings/$H_NEW_ID/postpone -H "Content-Type: application/json" -d '{
  "new_date":"2026-06-10","new_time_slot":"morning","reason":"证据未到位"
}'
echo ""

echo "=== D. Test lawyer conflict: same lawyer same slot ==="
# Schedule a 3rd case with lawyer law1, same slot as H1 (which uses law1)
LAWYER_CASE=$(curl -s -X POST $BASE/cases -H "Content-Type: application/json" -d '{
  "case_number":"(2026)刑初002","case_type":"criminal","title":"李四故意伤害案",
  "lawyers":[{"id":"law1","name":"李律师","phone":"13900000001","email":"l@b.com","firm":"正义所"}],
  "courtroom_size":"large"
}')
LC_ID=$(echo "$LAWYER_CASE" | JQ '.id')
curl -s -X POST $BASE/hearings -H "Content-Type: application/json" -d "{
  \"case_id\":\"$LC_ID\",\"judge_id\":\"$J_ID\",\"courtroom_id\":\"$CR3_ID\",
  \"date\":\"2026-07-10\",\"time_slot\":\"morning\",\"duration_min\":60
}"
echo ""
echo "=== DONE ==="