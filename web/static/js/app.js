document.getElementById('searchButton').addEventListener('click', async () => {
    const orderId = document.getElementById('orderIdInput').value;
    const resultDiv = document.getElementById('result');
    resultDiv.innerHTML = ''; // Очищаем предыдущий результат

    if (!orderId) {
        resultDiv.innerHTML = '<p class="result-error">Введите ID заказа.</p>';
        return;
    }

    try {
        const response = await fetch(`/order/${orderId}`);
        const data = await response.json();

        if (response.status === 404) {
            resultDiv.innerHTML = `<p class="result-error">Заказ с ID "${orderId}" не найден.</p>`;
        } else if (!response.ok) {
            resultDiv.innerHTML = `<p class="result-error">Ошибка сервера: ${data.message || response.statusText}</p>`;
        } else {
            resultDiv.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
        }
    } catch (error) {
        resultDiv.innerHTML = `<p class="result-error">Произошла ошибка при запросе: ${error.message}</p>`;
    }
});