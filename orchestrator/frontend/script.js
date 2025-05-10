document.getElementById('calculateForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const expression = document.getElementById('expression').value;

    fetch('/api/v1/calculate', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ expression: expression }),
    })
    .then(response => response.json())
    .then(data => {
        alert('Выражение отправлено на вычисление. ID: ' + data.id);
        fetchAllExpressions();
    })
    .catch(error => console.error('Ошибка:', error));
});

function fetchExpressionById() {
    const expressionId = document.getElementById('expressionId').value;
    if (!expressionId) {
        alert('Введите ID выражения');
        return;
    }

    fetch(`/api/v1/expressions/${expressionId}`)
    .then(response => {
        if (!response.ok) {
            throw new Error('Выражение не найдено');
        }
        return response.json();
    })
    .then(data => {
        const expr = data.expression || data;
        const resultDiv = document.getElementById('expressionResult');
        resultDiv.innerHTML = `
            <div class="alert alert-success">
                <strong>ID:</strong> ${expr.id}<br>
                <strong>Выражение:</strong> ${expr.expression}<br>
                <strong>Статус:</strong> ${expr.status}<br>
                <strong>Результат:</strong> ${expr.result || 'В процессе вычисления'}
            </div>
        `;
    })
    .catch(error => {
        console.error('Ошибка:', error);
        document.getElementById('expressionResult').innerHTML = `
            <div class="alert alert-danger">
                Не удалось найти выражение с ID ${expressionId}
            </div>
        `;
    });
}

function fetchAllExpressions() {
    fetch('/api/v1/expressions')
    .then(response => {
        if (!response.ok) {
            throw new Error('Ошибка при загрузке выражений');
        }
        return response.json();
    })
    .then(data => {
        const resultsList = document.getElementById('results');
        resultsList.innerHTML = '';

        const expressions = data.expressions || [];
        expressions.forEach(item => {
            const li = document.createElement('li');
            li.className = 'list-group-item';
            li.innerHTML = `
                <strong>ID:</strong> ${item.id}<br>
                <strong>Выражение:</strong> ${item.expression}<br>
                <strong>Статус:</strong> ${item.status}<br>
                <strong>Результат:</strong> ${item.result || 'В процессе вычисления'}
            `;
            resultsList.appendChild(li);
        });
    })
    .catch(error => {
        console.error('Ошибка:', error);
        const resultsList = document.getElementById('results');
        resultsList.innerHTML = '<li class="list-group-item text-danger">Ошибка при загрузке выражений</li>';
    });
}

function fetchTask() {
    fetch('/internal/task')
    .then(response => response.json())
    .then(data => {
        const task = data.task;
        const taskResultDiv = document.getElementById('taskResult');
        taskResultDiv.innerHTML = `
            <div class="alert alert-secondary">
                <strong>ID:</strong> ${task.id}<br>
                <strong>Аргумент 1:</strong> ${task.arg1}<br>
                <strong>Аргумент 2:</strong> ${task.arg2}<br>
                <strong>Операция:</strong> ${task.operation}<br>
                <strong>Время выполнения:</strong> ${task.operation_time} сек
            </div>
        `;
    })
    .catch(error => console.error('Ошибка:', error));
}

document.getElementById('sendResultForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const taskId = document.getElementById('taskId').value;
    const result = document.getElementById('taskResultValue').value;

    if (!taskId || !result) {
        alert("Пожалуйста, заполните все поля.");
        return;
    }

    const requestBody = {
        id: parseInt(taskId),
        result: parseFloat(result)
    };

    fetch('/internal/task', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("Ошибка при отправке результата");
        }
        return response.json();
    })
    .then(data => {
        alert('Результат успешно отправлен: ' + data.result);
    })
    .catch(error => {
        console.error('Ошибка:', error);
        alert('Ошибка при отправке результата: ' + error.message);
    });
});

function renderCharts(data) {
    const operationsCtx = document.getElementById('operationsChart').getContext('2d');

    new Chart(operationsCtx, {
        type: 'bar',
        data: {
            labels: Object.keys(data.operations),
            datasets: [{
                label: 'Количество операций',
                data: Object.values(data.operations),
                backgroundColor: [
                    'rgba(255, 99, 132, 0.2)',
                    'rgba(54, 162, 235, 0.2)',
                    'rgba(255, 206, 86, 0.2)',
                    'rgba(75, 192, 192, 0.2)',
                ],
                borderColor: [
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                ],
                borderWidth: 1
            }]
        },
        options: {
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

function fetchStatistics() {
    fetch('/api/v1/statistics')
        .then(response => response.json())
        .then(data => {
            console.log("Statistics data:", data);
            renderCharts(data);
        })
        .catch(error => console.error('Ошибка:', error));
}

setInterval(fetchStatistics, 5000);

fetchStatistics();