import http from 'k6/http';
import { sleep, check } from 'k6';
import { Trend } from 'k6/metrics';

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

let getTrend = new Trend('ACC_Query_Resp');

export let options = {
    stages: [
        { duration: '30s', target: 50 }, // Ramp up to 50 users over 30 seconds
        { duration: '1m', target: 50 },  // Stay at 50 users for 1 minute
        { duration: '10s', target: 0 },  // Ramp down to 0 users over 10 seconds
    ],
};

export default function () {
    // GET request
    let getRes = http.get(`http://localhost/api/v1/accounts/${getRandomInt(1, 4000)}`);
    check(getRes, {
        'GET status is 200': (r) => r.status === 200,
    });
    getTrend.add(getRes.timings.duration);


    sleep(1);
}
