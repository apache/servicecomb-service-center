## how to run dev environment?
### With docker:
```bash
cd dev
docker-compose
```
visit frontend at http://localhost:4200
#### change cp-frontend
Do nothing, change will be refreshed automatically.
### Without docker:
Start service-center first! Then:
```bash
cd cp-front
npm install
npm start -- --proxy-config=proxy.conf.json --host=0.0.0.0
```
visit frontend at http://localhost:4200
#### change cp-frontend
Do nothing, change will be refreshed automatically.
```
