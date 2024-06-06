import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  scenarios: {
    contacts: {
      executor: 'ramping-arrival-rate',
      preAllocatedVUs: 50,
      timeUnit: '1s',
      startRate: 50,
      stages: [
        { target: 200, duration: '30s' }, // linearly go from 50 iters/s to 200 iters/s for 30s
        { target: 500, duration: '0' }, // instantly jump to 500 iters/s
        { target: 500, duration: '10m' }, // continue with 500 iters/s for 10 minutes
      ],
    },
  },
};

export default function () {
  const data = { email: 'root@localhost', password: '' }
  let res = http.post('http://localhost:5680/api/v1/auth/token', data)

  check(res, { 'success login': (r) => r.status === 200 })
  let token = res.json('token')

  let machines = http.get('http://localhost:5680/api/v1/machines', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  check(machines, { 'success login': (r) => r.status === 200 })

  sleep(0.1)
}
