import http from 'k6/http';
export const options = {
    vus: 10,
    duration: '30s',
};
export function setup() {
    const url3 = 'http://127.0.0.1:30100/v4/default/registry/microservices/test3?force=1';
    http.del(url3);

    const url = 'http://127.0.0.1:30100/v4/default/registry/microservices';
    const payload = JSON.stringify({
        service: {
            serviceName: 'test3',
            serviceId: 'test3',
        }});

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    http.post(url, payload, params);
    for (let index = 0; index < 100; index++){
        const url2 = 'http://127.0.0.1:30100/v4/default/registry/microservices/test3/instances';
        const payload2 = JSON.stringify({
            instance: {
                hostName: "tian",
                endpoints: [
                    "ex of",
                    "labore"
                ],
                serviceId: "test3",
                properties: {},
                dataCenterInfo: {
                    name: "beijing",
                    region: "beijing",
                    availableZone: "az1"
                }
            }
        });

        const params2 = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        http.post(url2, payload2, params2);
    }

}
export default function () {
    const url = 'http://127.0.0.1:30100/v4/default/registry/instances?appId=default&serviceName=test3';
    http.get(url);
}

export function teardown(data) {
    const url = 'http://127.0.0.1:30100/v4/default/registry/microservices/test3?force=1';

    http.del(url);
}