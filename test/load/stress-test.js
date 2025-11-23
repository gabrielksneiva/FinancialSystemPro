import http from 'k6/http';
import { check, sleep } from 'k6';

// Teste de stress - aumenta carga até quebrar
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Aquecimento
    { duration: '5m', target: 200 },   // Carga crescente
    { duration: '2m', target: 300 },   // Mais carga
    { duration: '2m', target: 400 },   // Estresse
    { duration: '2m', target: 500 },   // Limite
    { duration: '10m', target: 0 },    // Recovery
  ],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // Health check simples
  const res = http.get(`${BASE_URL}/health`);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  sleep(0.5);
}
