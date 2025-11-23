import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Métricas customizadas
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const depositDuration = new Trend('deposit_duration');
const transferDuration = new Trend('transfer_duration');

// Configuração do teste
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp-up: 10 usuários
    { duration: '1m', target: 50 },    // Carga média: 50 usuários
    { duration: '2m', target: 100 },   // Carga alta: 100 usuários
    { duration: '1m', target: 50 },    // Ramp-down: 50 usuários
    { duration: '30s', target: 0 },    // Finalizar
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    'errors': ['rate<0.1'],                           // Taxa de erro < 10%
    'http_req_failed': ['rate<0.05'],                 // Falhas < 5%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Dados de teste
const testUsers = [];
for (let i = 0; i < 100; i++) {
  testUsers.push({
    email: `loadtest_user_${i}_${Date.now()}@example.com`,
    password: 'TestPassword123!',
  });
}

export function setup() {
  console.log('🚀 Starting load test against:', BASE_URL);
  
  // Criar alguns usuários de teste
  const users = [];
  for (let i = 0; i < 5; i++) {
    const user = testUsers[i];
    const res = http.post(`${BASE_URL}/api/users`, JSON.stringify(user), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (res.status === 200 || res.status === 201) {
      users.push(user);
    }
  }
  
  console.log(`✅ Setup completed. ${users.length} test users created.`);
  return { users };
}

export default function (data) {
  const user = data.users[Math.floor(Math.random() * data.users.length)];
  
  // Cenário 1: Login (40% das requests)
  if (Math.random() < 0.4) {
    testLogin(user);
  }
  // Cenário 2: Depósito (30% das requests)
  else if (Math.random() < 0.7) {
    testDeposit(user);
  }
  // Cenário 3: Transferência (20% das requests)
  else if (Math.random() < 0.9) {
    testTransfer(user);
  }
  // Cenário 4: Consulta de saldo (10% das requests)
  else {
    testBalance(user);
  }
  
  sleep(1); // Pausa entre requests
}

function testLogin(user) {
  const startTime = Date.now();
  
  const res = http.post(
    `${BASE_URL}/api/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  const success = check(res, {
    'login status 200': (r) => r.status === 200,
    'login has token': (r) => r.json('token') !== undefined,
  });
  
  errorRate.add(!success);
  loginDuration.add(Date.now() - startTime);
  
  return res.json('token');
}

function testDeposit(user) {
  const token = testLogin(user);
  if (!token) return;
  
  const startTime = Date.now();
  
  const payload = {
    amount: Math.floor(Math.random() * 1000) + 10, // 10-1010
    callback_url: 'https://webhook.site/test',
  };
  
  const res = http.post(
    `${BASE_URL}/api/deposit`,
    JSON.stringify(payload),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  const success = check(res, {
    'deposit status 200': (r) => r.status === 200 || r.status === 201,
  });
  
  errorRate.add(!success);
  depositDuration.add(Date.now() - startTime);
}

function testTransfer(user) {
  const token = testLogin(user);
  if (!token) return;
  
  const startTime = Date.now();
  
  const payload = {
    to_user_id: 'recipient-uuid-here', // Ajustar conforme necessário
    amount: Math.floor(Math.random() * 100) + 1,
  };
  
  const res = http.post(
    `${BASE_URL}/api/transfer`,
    JSON.stringify(payload),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  const success = check(res, {
    'transfer status ok': (r) => r.status === 200 || r.status === 400, // 400 pode ser falta de saldo
  });
  
  errorRate.add(!success && res.status !== 400);
  transferDuration.add(Date.now() - startTime);
}

function testBalance(user) {
  const token = testLogin(user);
  if (!token) return;
  
  const res = http.get(`${BASE_URL}/api/balance`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  
  const success = check(res, {
    'balance status 200': (r) => r.status === 200,
    'balance has value': (r) => r.json('balance') !== undefined,
  });
  
  errorRate.add(!success);
}

export function teardown(data) {
  console.log('🧹 Cleaning up test data...');
  // Aqui você pode limpar os dados de teste se necessário
}
