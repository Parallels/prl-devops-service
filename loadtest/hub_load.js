// Usage:
// 1. Allow multiple WebSocket connections from the same IP:
//    - build the binary with `-tags=debug`, or
//    - temporarily change src/serviceprovider/eventEmitter/connection_limit_release.go to return true.
// 2. Start the API locally (http://localhost:5680) and run `prldevops test event-load` to stream events.
// 3. Execute: k6 run loadtest/hub_load.js

import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const latencyTrend = new Trend('message_latency');
const messageLossCounter = new Counter('message_loss');
const connectionErrors = new Counter('connection_errors');
const messagesReceived = new Counter('messages_received');
const pingsSent = new Counter('pings_sent');
const pongsReceived = new Counter('pongs_received');

export const options = {
    scenarios: {
        hub_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 400 },
                { duration: '2m', target: 400 },
                { duration: '30s', target: 0 },
            ],
            gracefulStop: '30s',
        },
    },
};

export default function () {
    const data = { email: 'root@localhost', password: '' };
    const authRes = http.post('http://localhost:5680/api/v1/auth/token', data);
    check(authRes, { 'status is 200': (r) => r.status === 200 });

    const token = authRes.json('token');
    const url = `ws://localhost:5680/api/v1/ws/subscribe?event_types=pdfm,health`;
    const params = {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { my_tag: 'hub_load' }
    };

    // 2. Connect to WebSocket
    const response = ws.connect(url, params, function (socket) {
        let lastSeq = -1;
        let pingInterval;
        let localPongCount = 0;
        let localPingCount = 0;

        socket.on('open', function open() {
            // Send first ping immediately
            const sendPing = function () {
                const pingMsg = {
                    type: 'health',
                    message: 'ping'
                };
                socket.send(JSON.stringify(pingMsg));
                localPingCount++;
                pingsSent.add(1);
            };

            // Send first ping right away
            sendPing();

            // Then send pings every 5 seconds
            pingInterval = setInterval(sendPing, 5000);
        });

        socket.on('message', function (message) {
            const msg = JSON.parse(message);

            // Track pong responses
            if (msg.event_type === 'health' && msg.message === 'pong') {
                localPongCount++;
                pongsReceived.add(1);
            }

            // Track PDFM VM events with bigger payload
            if (msg.event_type === 'pdfm' && msg.body && msg.body.new_vm && msg.body.new_vm.ID) {
                messagesReceived.add(1);

                // Parse seq and timestamp from ID format: "seq-{seq}-ts-{timestamp}"
                const vmId = msg.body.new_vm.ID;
                const parts = vmId.split('-');
                if (parts.length >= 4 && parts[0] === 'seq' && parts[2] === 'ts') {
                    const seq = parseInt(parts[1]);
                    const sentTime = parseInt(parts[3]);
                    const now = new Date().getTime() * 1000000; // ns

                    // Calculate latency (ms)
                    const latencyMs = (now - sentTime) / 1000000;
                    if (latencyMs > 0) {
                        latencyTrend.add(latencyMs);
                    }

                    // Check for loss
                    if (lastSeq !== -1) {
                        const diff = seq - lastSeq;
                        if (diff > 1) {
                            messageLossCounter.add(diff - 1);
                        }
                    }
                    lastSeq = seq;
                }
            }
        });

        socket.on('close', function () {
            if (pingInterval) {
                clearInterval(pingInterval);
            }
        });

        socket.on('error', function (e) {
            connectionErrors.add(1);
            console.log('error: ', e.error());
        });

        // Keep connection open for a while
        socket.setTimeout(function () {
            if (pingInterval) {
                clearInterval(pingInterval);
            }
            socket.close();
        }, 30000); // 30s session per VU iteration
    });

    check(response, { 'status is 101': (r) => r && r.status === 101 });
    sleep(1);
}
