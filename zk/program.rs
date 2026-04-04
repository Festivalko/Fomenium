// zk/program.rs
// Это программа, которую SP1 превратит в ZK доказательство
#![no_main]
sp1_zkvm::entrypoint!(main);

use sp1_zkvm::io;
use sha2::{Sha256, Digest};

// Структура платежа для ZK
#[derive(serde::Serialize, serde::Deserialize, Debug)]
struct Payment {
    from: u32,
    to: u32,
    amount: u64,
    nonce: u64,
}

// Структура батча для ZK
#[derive(serde::Serialize, serde::Deserialize, Debug)]
struct BatchProof {
    batch_id: String,
    payments: Vec<Payment>,
    initial_balances: Vec<(u32, u64)>,
    final_balances: Vec<(u32, u64)>,
    total_amount: u64,
    fee: u64,
}

pub fn main() {
    // Читаем входные данные от Go приложения
    let batch: BatchProof = io::read();
    
    // Создаем карту балансов
    let mut balances = std::collections::HashMap::new();
    for (account, balance) in &batch.initial_balances {
        balances.insert(*account, *balance);
    }
    
    // Проверяем каждый платеж
    let mut total_amount = 0u64;
    let mut final_balances = balances.clone();
    
    for payment in &batch.payments {
        // Проверяем, что у отправителя достаточно средств
        let from_balance = balances.get(&payment.from).expect("Account not found");
        assert!(*from_balance >= payment.amount, 
                "Insufficient balance for account {}", payment.from);
        
        // Обновляем балансы
        *final_balances.get_mut(&payment.from).unwrap() -= payment.amount;
        *final_balances.entry(payment.to).or_insert(0) += payment.amount;
        
        total_amount += payment.amount;
    }
    
    // Проверяем, что финальные балансы соответствуют
    for (account, expected_balance) in &batch.final_balances {
        let actual_balance = final_balances.get(account)
            .expect("Account not found in final state");
        assert_eq!(actual_balance, expected_balance,
                   "Balance mismatch for account {}", account);
    }
    
    // Проверяем комиссию (0.1% от суммы)
    let expected_fee = total_amount / 1000;
    assert_eq!(batch.fee, expected_fee,
               "Fee mismatch: expected {}, got {}", expected_fee, batch.fee);
    
    // Вычисляем хэш результата (будет частью proof)
    let mut hasher = Sha256::new();
    hasher.update(batch.batch_id.as_bytes());
    for payment in &batch.payments {
        hasher.update(&payment.from.to_le_bytes());
        hasher.update(&payment.to.to_le_bytes());
        hasher.update(&payment.amount.to_le_bytes());
        hasher.update(&payment.nonce.to_le_bytes());
    }
    let result_hash = hasher.finalize();
    
    // Выводим результат (будет доступен в proof.public_values)
    io::write(&result_hash);
    
    // Выводим финальные балансы (для верификации)
    for (account, balance) in &final_balances {
        io::write(&account);
        io::write(&balance);
    }
}