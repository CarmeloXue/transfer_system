import http from 'k6/http';
import { sleep, check } from 'k6';
import { Trend } from 'k6/metrics';

let postTrend = new Trend('ACC_CREATE_RESP');

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}


export let options = {
    stages: [
        { duration: '30s', target: 50 }, // Ramp up to 50 users over 30 seconds
        { duration: '1m', target: 50 },  // Stay at 50 users for 1 minute
        { duration: '10s', target: 0 },  // Ramp down to 0 users over 10 seconds
    ],
};

export default function () {

    // POST request


    for (let i = 1; i < 4000; i++) {
        let postPayload = JSON.stringify({
            account_id: i,
            initial_balance: '123123.1'
        });
        let postParams = {
            headers: { 'Content-Type': 'application/json' },
        };
        let postRes = http.post('http://localhost/api/v1/accounts', postPayload, postParams);
        check(postRes, {
            'POST status is 201': (r) => r.status === 201,
        });
        postTrend.add(postRes.timings.duration);
    }



    sleep(1);
}
