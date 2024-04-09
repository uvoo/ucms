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
```
