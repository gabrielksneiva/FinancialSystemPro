import http from 'k6/http';
import { check } from 'k6';

// Teste de soak - carga constante por longo período
export const options = {
  stages: [
    { duration: '5m', target: 100 },   // Ramp-up
    { duration: '4h', target: 100 },   // Carga constante por 4 horas
    { duration: '5m', target: 0 },     // Ramp-down
  ],
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${BASE_URL}/health`);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
