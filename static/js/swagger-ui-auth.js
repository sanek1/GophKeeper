// Скрипт для добавления кнопки авторизации в Swagger UI
window.addEventListener('load', function() {
    // Дожидаемся полной загрузки Swagger UI
    setTimeout(function() {
        // Создаем кнопку авторизации
        const authButton = document.createElement('button');
        authButton.className = 'btn authorize unlocked';
        authButton.innerText = 'Authorize';
        
        // Добавляем стили
        authButton.style.backgroundColor = '#49cc90';
        authButton.style.color = '#fff';
        authButton.style.border = 'none';
        authButton.style.padding = '10px 15px';
        authButton.style.borderRadius = '4px';
        authButton.style.cursor = 'pointer';
        authButton.style.position = 'absolute';
        authButton.style.right = '20px';
        authButton.style.top = '70px';
        authButton.style.zIndex = '1000';
        
        // Добавляем обработчик события для показа диалога авторизации
        authButton.addEventListener('click', function() {
            // Показываем модальное окно для ввода токена
            const modal = document.createElement('div');
            modal.style.position = 'fixed';
            modal.style.top = '0';
            modal.style.left = '0';
            modal.style.width = '100%';
            modal.style.height = '100%';
            modal.style.backgroundColor = 'rgba(0,0,0,0.5)';
            modal.style.display = 'flex';
            modal.style.justifyContent = 'center';
            modal.style.alignItems = 'center';
            modal.style.zIndex = '2000';
            
            const modalContent = document.createElement('div');
            modalContent.style.backgroundColor = '#fff';
            modalContent.style.padding = '20px';
            modalContent.style.borderRadius = '5px';
            modalContent.style.width = '400px';
            
            const header = document.createElement('h3');
            header.innerText = 'Авторизация';
            header.style.marginTop = '0';
            
            const tokenInput = document.createElement('input');
            tokenInput.type = 'text';
            tokenInput.placeholder = 'Введите JWT токен';
            tokenInput.style.width = '100%';
            tokenInput.style.padding = '8px';
            tokenInput.style.marginBottom = '10px';
            tokenInput.style.boxSizing = 'border-box';
            
            const buttonContainer = document.createElement('div');
            buttonContainer.style.display = 'flex';
            buttonContainer.style.justifyContent = 'space-between';
            
            const saveButton = document.createElement('button');
            saveButton.innerText = 'Сохранить';
            saveButton.style.padding = '8px 15px';
            saveButton.style.backgroundColor = '#49cc90';
            saveButton.style.color = '#fff';
            saveButton.style.border = 'none';
            saveButton.style.borderRadius = '4px';
            saveButton.style.cursor = 'pointer';
            
            const cancelButton = document.createElement('button');
            cancelButton.innerText = 'Отмена';
            cancelButton.style.padding = '8px 15px';
            cancelButton.style.backgroundColor = '#f93e3e';
            cancelButton.style.color = '#fff';
            cancelButton.style.border = 'none';
            cancelButton.style.borderRadius = '4px';
            cancelButton.style.cursor = 'pointer';
            
            buttonContainer.appendChild(cancelButton);
            buttonContainer.appendChild(saveButton);
            
            modalContent.appendChild(header);
            modalContent.appendChild(tokenInput);
            modalContent.appendChild(buttonContainer);
            modal.appendChild(modalContent);
            
            document.body.appendChild(modal);
            
            // Обработчик для сохранения токена
            saveButton.addEventListener('click', function() {
                const token = tokenInput.value.trim();
                
                if (token) {
                    // Сохраняем токен в localStorage
                    localStorage.setItem('swagger_ui_bearer_token', token);
                    
                    // Обновляем все запросы с новым токеном
                    const authHeader = token.startsWith('Bearer ') ? token : `Bearer ${token}`;
                    
                    // Применяем токен ко всем операциям
                    const operations = document.querySelectorAll('.opblock');
                    operations.forEach(function(operation) {
                        if (operation.classList.contains('opblock-get') || 
                            operation.classList.contains('opblock-post') || 
                            operation.classList.contains('opblock-put') || 
                            operation.classList.contains('opblock-delete')) {
                            
                            // Находим запрос этой операции
                            const opId = operation.getAttribute('id');
                            if (opId) {
                                const request = window.ui.fn.curlify({
                                    operationId: opId.replace('operations-', ''),
                                    server: '',
                                    baseUrl: '',
                                    url: '',
                                });
                                
                                if (request && request.headers) {
                                    request.headers.Authorization = authHeader;
                                }
                            }
                        }
                    });
                    
                    authButton.classList.remove('unlocked');
                    authButton.classList.add('locked');
                    authButton.innerText = 'Authorized';
                    
                    // Показываем сообщение об успешной авторизации
                    alert('Токен успешно сохранен. Вы авторизованы.');
                }
                
                // Закрываем модальное окно
                document.body.removeChild(modal);
            });
            
            // Обработчик для закрытия модального окна
            cancelButton.addEventListener('click', function() {
                document.body.removeChild(modal);
            });
        });
        
        // Проверяем, есть ли сохраненный токен
        const savedToken = localStorage.getItem('swagger_ui_bearer_token');
        if (savedToken && savedToken.trim() !== '') {
            authButton.classList.remove('unlocked');
            authButton.classList.add('locked');
            authButton.innerText = 'Authorized';
            
            // Применяем токен ко всем операциям
            const authHeader = savedToken.startsWith('Bearer ') ? savedToken : `Bearer ${savedToken}`;
            
            // Добавляем обработчик для всех запросов
            const originalFetch = window.fetch;
            window.fetch = function(url, options) {
                if (options && url.includes('/api/')) {
                    if (!options.headers) {
                        options.headers = {};
                    }
                    options.headers.Authorization = authHeader;
                }
                return originalFetch.apply(this, arguments);
            };
        }
        
        // Добавляем кнопку на страницу
        document.body.appendChild(authButton);
    }, 1000); // Задержка для полной загрузки Swagger UI
}); 