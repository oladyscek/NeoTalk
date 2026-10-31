self.addEventListener('push', function(event) {
    let data = {};
    try {
        data = event.data.json();
    } catch (e) {
        data = { title: 'NeoTalk', body: 'New message' };
    }

    const options = {
        body: data.body || 'New message',
        icon: '/favicon.png',
        badge: '/favicon.png',
        vibrate: [100, 50, 100],
        data: { url: '/' } // Куда перейти при клике
    };

    event.waitUntil(
        self.registration.showNotification(data.title || 'NeoTalk', options)
    );
});

// Обработка клика по уведомлению — открываем вкладку
self.addEventListener('notificationclick', function(event) {
    event.notification.close();
    event.waitUntil(
        clients.matchAll({ type: 'window' }).then(windowClients => {
            if (windowClients.length > 0) {
                return windowClients[0].focus();
            }
            return clients.openWindow('/');
        })
    );
});