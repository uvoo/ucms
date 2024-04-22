sqlite3 ucms.db "insert into country_code_rules (code, action, priority) values ('Private', 'allow', 10)"
sqlite3 ucms.db "insert into country_code_rules (code, action, priority) values ('US', 'allow', 11)"
./curl.sh

