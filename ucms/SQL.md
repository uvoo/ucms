```
update pages set content = '<h1>Hello mdboostrap!</h1>Smile 2<i class="fas fa-smile"></i></br></br><button type="button" class="btn btn-primary" data-mdb-ripple-init>Button</button>
' where id = 3;

INSERT INTO pages (title, content) VALUES ('My MDBootstrap', '<h1>Hello mdboostrap!</h1>Smile 2<i class="fas fa-smile"></i></br></br><button type="button" class="btn btn-primary" data-mdb-ripple-init>Button</button>')
insert into pages (template, title, content) VALUES ('markdown', 'My markdown', '## Heading\n- one - two - three')


insert into pages (template, title, content) VALUES ('markdown', 'My markdown', '
 # header

Sample text.

[link](http://example.com)');


update pages set name = 'one' where id = 1;

sqlite3 ucms.db "select * from pages"



sqlite3 ucms.db "insert into country_code_rules (code, action) values ('US', 'allow')"
sqlite3 ucms.db "insert into fw_rules (src_ip_net, action) values ('10.0.0.0/8', 'deny')"
sqlite3 ucms.db "insert into fw_rules (src_ip_net, action) values ('10.0.0.0/8', 'allow')"
sqlite3 ucms.db "insert into fw_rules (src_ip_net, action, priority) values ('10.0.0.0/8', 'allow', 1);"

DROP INDEX idx_fw_rules_src_ip_net

sqlite3 ucms.db "select * from fw_rules"
sqlite3 ucms.db ".schema"
sqlite3 ucms.db "delete from fw_rules"
sqlite3 ucms.db "insert into country_code_rules (code, action, priority) values ('Private', 'allow', 20)"
e', 'allow', 20)"
sqlite3 ucms.db "insert into country_code_rules (code, action, priority) values ('US', 'allow', 21)"
sqlite3 ucms.db "select * from country_code_rules"
sqlite3 ucms.db "delete from country_code_rules"

sqlite3 ucms.db "insert into fw_rules (src_ip_net, action, priority) values ('10.1.1.0/24', 'allow', 10)"
sqlite3 ucms.db "insert into fw_rules (src_ip_net, action, priority) values ('10.1.1.0/24', 'drop', 9)"


sqlite3 ucms.db "delete from fw_rules" && sqlite3 ucms.db "delete from country_code_rules"
sqlite3 ucms.db "insert into fw_rules (src_ip_net, action, priority) values ('10.1.1.0/24', 'allow', 10)"


sqlite3 ucms.db "insert into country_code_rules (code, action, priority) values ('Private', 'allow', 20)"
sqlite3 ucms.db "INSERT INTO users (username, password, name, email) VALUES ('foo', 'bar', 'foo test', 'foo@uvoo.io')"
sqlite3 ucms.db "delete from users where name ='foo'"
```
