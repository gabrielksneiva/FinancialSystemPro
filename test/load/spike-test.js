import http from 'k6/http';
import { check, sleep } from 'k6';

// Teste de spike - picos repentinos de carga
export const options = {
  stages: [
    { duration: '10s', target: 10 },    // Carga normal
    { duration: '1m', target: 1000 },   // SPIKE!
    { duration: '10s', target: 10 },    // Volta ao normal
    { duration: '10s', target: 10 },    // Estabiliza
    { duration: '1m', target: 1000 },   // Outro SPIKE
    { duration: '10s', target: 10 },    // Recuperação
  ],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const endpoints = ['/health', '/ready', '/alive'];
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  
  const res = http.get(`${BASE_URL}${endpoint}`);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'no errors': (r) => !r.error,
  });
  
  sleep(0.1);
}
