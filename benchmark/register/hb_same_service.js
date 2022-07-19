import http from 'k6/http';
export const options = {
    vus: 10,
    duration: '30s',
};
export function setup() {
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

    const url2 = 'http://127.0.0.1:30100/v4/default/registry/microservices/test3/instances';
    const payload2 = JSON.stringify({
        instance: {
            hostName: "tian",
            endpoints: [
                "ex of",
                "labore"
            ],
            instanceId: "1",
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
export default function () {
    const url = 'http://127.0.0.1:30100/v4/default/registry/heartbeats';
    const payload = JSON.stringify({
        Instances: [
            {
                serviceId: "test3",
                instanceId: "1"
            }
        ]
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    http.post(url, payload, params);
}