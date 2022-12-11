# Goto

Interactive shell for quick SSH access to application servers/environments via simple configuration file.

### Configuration example

```yaml
version: 1

servers:
  server_1: &server_1
    username: 'joe'
    port: 22
    environments:
      stg:
        hosts: ['i1.stg.server1.example', 'i2.stg.server1.example']
      prod:
        hosts: ['i1.prod.server1.example', 'i2.prod.server1.example']
  server_2: &server_2
    username: 'joe'
    port: 22
    environments:
      stg:
        hosts: ['i1.stg.server2.example', 'i2.stg.server2.example']
      prod:
        hosts: ['i1.prod.server2.example', 'i2.prod.server2.example']

applications:
  app_1:
    server: *server_2
    username: 'org-app-1'
    path: '/home/org-app-1/deployer/current'
    scripts:
      - &cache-clear {name: 'cache:clear', exec: 'bin/php bin/console cache:clear'}
  app_2:
    server: *server_1
    username: 'org-app-2'
    path: '/home/org-app-2/deployer/current'
    scripts:
      - *cache-clear
  app_3:
    server: *server_2
    username: 'org-app-3'
    path: '/home/org-app-3/deployer/current'
    scripts:
      - *cache-clear
      - { name: 'generate:file', exec: 'bin/php bin/console generate:file' }
      - { name: 'killall-php', exec: 'killall php8.1' }
  app_4:
    server: *server_1
    username: 'org-app-4'
    path: '/home/org-app-4/deployer/current'
```