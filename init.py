ports = ['3001', '3002', '3003']
for port in ports:
    db = open('db:'+port+'.json', 'w')
    log = open('log:'+port+'.json', 'w')
    db.write('{}')
    log.write('[]')
    db.close()
    log.close()

