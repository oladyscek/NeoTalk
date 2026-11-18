// ================== I18N TRANSLATIONS ==================
const translations = {
  en: {
    conn_lost: 'CONNECTION LOST. WAITING FOR NETWORK...', name: 'Name', password: 'Password', repeat_pass: 'Repeat Password',
    log_in: 'Log In', register: 'Register', back_to_login: 'Back to Log In', create: 'Create...', new_chat: 'New Chat',
    new_group: 'New Group', back: 'Back', username: 'Username', group_name: 'Group Name', add_member: 'Add Member (Enter)',
    settings: 'Settings', profile: 'Profile', log_out: 'Log Out', change_lang: 'Change Language', enable_notifs: 'Enable Notifications',
    coming_soon: 'Coming soon..', chats: 'Chats', message: 'Message...', chat_actions: 'Chat Actions', delete_chat: 'Delete Chat',
    leave_group: 'Leave Group', rename_contact: 'Rename Contact', reset_name: 'Reset Name', members: 'Members', no_members: 'No members found.',
    new_name: 'New Name', enter_name: 'Enter name...', save: 'Save', edit_message: 'Edit Message', error: 'Error', ok: 'OK',
    edit: 'Edit', copy: 'Copy', forward: 'Forward', delete: 'Delete', err_login: 'Login: ', err_reg: 'Registration: ', err_create: 'Create error: ',
    err_load_members: 'Failed to load members', err_rename: 'Failed to rename', err_reset_name: 'Failed to reset name',
    err_del_msg: 'Failed to delete message.', err_del_chat: 'Failed to delete chat', err_copy: 'Failed to copy messages',
    forward_soon: 'Message forwarding will be available in future updates!', msg_too_long: 'Message is too long!',
    msg_slow_down: 'Too many messages! Please wait a second.', msg_server_err: 'Server error while sending.',
    notifs_enabled: 'Notifications enabled', no_permission: 'No permission', disconnected: 'Disconnected: Account opened in another tab/device.',
    left_group: 'left the group'
  },
  ru: {
    conn_lost: 'СОЕДИНЕНИЕ ПОТЕРЯНО. ОЖИДАНИЕ СЕТИ...', name: 'Имя', password: 'Пароль', repeat_pass: 'Повторите пароль',
    log_in: 'Войти', register: 'Регистрация', back_to_login: 'Назад ко входу', create: 'Создать...', new_chat: 'Новый чат',
    new_group: 'Новая группа', back: 'Назад', username: 'Имя пользователя', group_name: 'Название группы', add_member: 'Добавить (Enter)',
    settings: 'Настройки', profile: 'Профиль', log_out: 'Выйти', change_lang: 'Сменить язык', enable_notifs: 'Включить уведомления',
    coming_soon: 'Скоро...', chats: 'Чаты', message: 'Сообщение...', chat_actions: 'Действия с чатом', delete_chat: 'Удалить чат',
    leave_group: 'Покинуть группу', rename_contact: 'Переименовать контакт', reset_name: 'Сбросить имя', members: 'Участники', no_members: 'Участники не найдены.',
    new_name: 'Новое имя', enter_name: 'Введите имя...', save: 'Сохранить', edit_message: 'Изменить сообщение', error: 'Ошибка', ok: 'ОК',
    edit: 'Изменить', copy: 'Копировать', forward: 'Переслать', delete: 'Удалить', err_login: 'Ошибка входа: ', err_reg: 'Ошибка регистрации: ', err_create: 'Ошибка создания: ',
    err_load_members: 'Не удалось загрузить участников', err_rename: 'Не удалось переименовать', err_reset_name: 'Не удалось сбросить имя',
    err_del_msg: 'Не удалось удалить сообщение.', err_del_chat: 'Не удалось удалить чат', err_copy: 'Не удалось скопировать сообщения',
    forward_soon: 'Пересылка сообщений появится в будущих обновлениях!', msg_too_long: 'Сообщение слишком длинное!',
    msg_slow_down: 'Слишком много сообщений! Подождите секунду.', msg_server_err: 'Ошибка сервера при отправке.',
    notifs_enabled: 'Уведомления включены', no_permission: 'Нет разрешения', disconnected: 'Отключено: Аккаунт открыт в другой вкладке/устройстве.',
    left_group: 'покинул(а) группу'
  }
};

let currentLang = localStorage.getItem('lang') || 'en';

function t(key) {
  return translations[currentLang][key] || translations['en'][key] || key;
}

function applyTranslations() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    el.textContent = t(el.getAttribute('data-i18n'));
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    el.placeholder = t(el.getAttribute('data-i18n-placeholder'));
  });
}

applyTranslations();

// ================== APPLICATION STATE ==================
let myName = '';
let currentChatId = null;
let lastAppendedDate = null;
const chatsById = {};
let groupMembers = [];
let customNamesMap = {}; 

let selectedMsgIds = new Set();
let touchTimer = null;
let isLongPress = false;
let chatTouchTimer = null;
let selectedChatIdForAction = null;

let oldestMessageId = null;
let hasMoreHistory = true;
let isLoadingHistory = false;

let outbox = [];
let openPopoverMsgId = null;

let pendingEditMsgId = null;
let pendingEditChatId = null;
let pendingEditOldText = null;
let pendingEditNewText = null;
let pendingRenameId = null;
let pendingRenameName = null;

// Drag selection state
let mouseDownTarget = null;
let dragStartX = 0;
let dragStartY = 0;
let isDragging = false;
let wasDragging = false;
let dragRows = [];

// ================== HELPERS ==================
function getDisplayName(originalName) {
  return customNamesMap[originalName] || originalName;
}

function showToast(msg) {
  const toast = document.getElementById('toast');
  toast.textContent = msg;
  toast.classList.add('show');
  setTimeout(() => toast.classList.remove('show'), 3000);
}

function parseDate(timeStr) {
  if (!timeStr) return new Date();
  if (timeStr.includes('T')) return new Date(timeStr);
  return new Date(timeStr.replace(' ', 'T') + 'Z');
}
function formatMsgTime(timeStr) { return parseDate(timeStr).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }); }
function formatDateDivider(timeStr) { return parseDate(timeStr).toLocaleDateString(currentLang === 'ru' ? 'ru-RU' : 'en-US', { day: 'numeric', month: 'long' }); }
function getDateStr(timeStr) {
  const d = parseDate(timeStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

// ================== NAVIGATION ==================
function navigate(view, data = {}) {
  if (history.state && history.state.view === view && history.state.chatId === data.chatId) return;
  history.pushState({ view, ...data }, '');
  renderView(history.state);
}

window.addEventListener('popstate', (e) => {
  if (e.state) renderView(e.state);
  else {
    if (localStorage.getItem('token')) { history.replaceState({ view: 'chats' }, ''); renderView({ view: 'chats' }); }
    else { history.replaceState({ view: 'login' }, ''); renderView({ view: 'login' }); }
  }
});

function renderView(state) {
  hideAllScreens();
  document.body.classList.remove('show-mobile-chat');
  closeMsgPopover();
  closeDotsMenu();

  if (state.view === 'chats') {
    currentChatId = null;
    chatTitle.textContent = '';
    headerAvatar.textContent = '';
    sendInput.value = '';
    messagesList.innerHTML = '';
    document.querySelectorAll('.chat-list li').forEach(x => x.classList.remove('active'));
    document.querySelector('.chat').classList.add('empty-chat');
    clearSelection();
  }

  switch(state.view) {
    case 'login': loginScreen.style.display = 'flex'; break;
    case 'reg': regScreen.style.display = 'flex'; break;
    case 'chats': chatsScreen.style.display = 'block'; break;
    case 'chat':
      document.querySelector('.chat').classList.remove('empty-chat');
      chatsScreen.style.display = 'block';
      document.body.classList.add('show-mobile-chat');
      if (currentChatId !== state.chatId) {
        const li = chatList.querySelector(`li[data-chat-id="${state.chatId}"]`);
        if (li) {
          document.querySelectorAll('.chat-list li').forEach(x => x.classList.remove('active'));
          li.classList.add('active');
          currentChatId = state.chatId;
          const chat = chatsById[currentChatId];
          if (chat) {
            chatTitle.textContent = chat.chatName;
            headerAvatar.textContent = chat.chatName.charAt(0).toUpperCase();
          }
          oldestMessageId = null; hasMoreHistory = true; isLoadingHistory = false; lastAppendedDate = null;
          messagesList.innerHTML = ''; clearSelection();
          clearUnreadUI(currentChatId);
          send({ type: 'chatRead', chatId: currentChatId });
          if (chat) send({ type: 'getHistory', chatId: currentChatId, beforeId: 0 });
        }
      }
      break;
    case 'settings': settingsScreen.style.display = 'flex'; break;
    case 'new-choice': newChoiceScreen.style.display = 'flex'; break;
    case 'new-chat': newChatScreen.style.display = 'flex'; newChatInput.focus(); break;
    case 'new-group': newGroupScreen.style.display = 'flex'; newGroupName.focus(); break;
    case 'coming-soon': soonScreen.style.display = 'flex'; break;
  }
}

function transitionToChats() {
  history.replaceState({ view: 'chats' }, '');
  renderView({ view: 'chats' });
}

// ================== WEBSOCKET ==================
let ws;
let reconnectTimeout = null;
const connectionStatus = document.getElementById('connection-status');

function connectWebSocket() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return;

  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  ws = new WebSocket(`${proto}://${location.host}/ws`);

  ws.onopen = () => {
    connectionStatus.style.display = 'none';
    const savedToken = localStorage.getItem('token');
    if (savedToken) send({ type: 'auth', token: savedToken });
  };

  ws.onclose = (e) => {
    if (e.code === 4000) {
      showError(t('disconnected'));
      setTimeout(() => location.reload(), 2000);
      return;
    }
    connectionStatus.style.display = 'block';

    const pendingEls = messagesList.querySelectorAll('.own.pending:not([data-msg-id])');
    pendingEls.forEach(el => {
      const cid = parseInt(el.dataset.chatId);
      const text = el.dataset.originalText;
      const timeStr = el.dataset.time;
      const chat = chatsById[cid];
      if (chat) outbox.push({ chatId: cid, isGroup: chat.isGroup, to: chat.peerName || chat.chatName, text: text, timeStr: timeStr });
      el.remove();
    });

    clearTimeout(reconnectTimeout);
    reconnectTimeout = setTimeout(connectWebSocket, 3000);
  };

  ws.onerror = (e) => console.error('ws error', e);

  ws.onmessage = (e) => {
    let p;
    try { p = JSON.parse(e.data); } catch { return; }
    console.log('<<', p);

    switch (p.type) {
      case 'authStatus':
        if (p.answer === 'ok') {
          const wasLoggedIn = myName !== '';
          myName = p.name;
          document.documentElement.classList.remove('auth-pending');
          send({ type: 'getChats' });
          send({ type: 'getCustomNames' });
          if (!wasLoggedIn) transitionToChats();
          flushOutbox();
        } else {
          localStorage.removeItem('token');
          document.documentElement.classList.remove('auth-pending');
          showLogin();
        }
        break;
      case 'loginStatus':
        if (p.answer === 'ok') {
          myName = loginNameInput.value.trim();
          localStorage.setItem('token', p.token);
          document.documentElement.classList.remove('auth-pending');
          send({ type: 'getChats' });
          send({ type: 'getCustomNames' });
          transitionToChats();
          flushOutbox();
        } else showError(t('err_login') + p.answer);
        break;
      case 'regStatus':
        if (p.answer === 'ok') {
          myName = regNameInput.value.trim();
          localStorage.setItem('token', p.token);
          document.documentElement.classList.remove('auth-pending');
          send({ type: 'getChats' });
          send({ type: 'getCustomNames' });
          transitionToChats();
          flushOutbox();
        } else showError(t('err_reg') + p.answer);
        break;
      case 'getCustomNamesStatus':
        if (p.answer === 'ok') {
          customNamesMap = p.data || {};
          if (currentChatId) renderHistory([]);
          send({type: 'getChats'});
        }
        break;
      case 'getChatStatus':
        if (p.answer === 'ok') renderChatList(p.data || []);
        break;
      case 'getHistoryStatus':
        if (p.answer === 'ok') {
          hasMoreHistory = !!p.hasMore;
          renderHistory(p.data || []);
          isLoadingHistory = false;
        } else isLoadingHistory = false;
        break;
      case 'getGroupMembersStatus':
        if (p.answer === 'ok') renderGroupMembers(p.data);
        else showError(t('err_load_members'));
        break;
      case 'makeCustomNameStatus':
        if (p.answer === 'ok') {
          if (pendingRenameId && pendingRenameName) {
            if (chatsById[pendingRenameId]) chatsById[pendingRenameId].chatName = pendingRenameName;
            const targetName = chatsById[pendingRenameId].peerName;
            if (targetName) customNamesMap[targetName] = pendingRenameName;
            const chatLi = chatList.querySelector(`li[data-chat-id="${pendingRenameId}"]`);
            if (chatLi) {
              chatLi.querySelector('.chat-name').textContent = pendingRenameName;
              chatLi.querySelector('.chat-avatar').textContent = pendingRenameName.charAt(0).toUpperCase();
            }
            if (pendingRenameId === currentChatId) {
              chatTitle.textContent = pendingRenameName;
              headerAvatar.textContent = pendingRenameName.charAt(0).toUpperCase();
              const profName = document.getElementById('profile-name');
              if (profName) profName.textContent = pendingRenameName;
              const profAva = document.getElementById('profile-avatar');
              if (profAva) profAva.textContent = pendingRenameName.charAt(0).toUpperCase();
            }
            pendingRenameId = null; pendingRenameName = null;
            send({type: 'getChats'});
          }
          renameModal.style.display = 'none';
          profileModal.style.display = 'none';
        } else showError(t('err_rename'));
        break;
      case 'delCustomNameStatus':
        if (p.answer === 'ok') {
          renameModal.style.display = 'none';
          profileModal.style.display = 'none';
          send({ type: 'getCustomNames' });
        } else showError(t('err_reset_name'));
        break;
      case 'incomedMsg':
        handleIncomingMsg(p);
        break;
      case 'newChatStatus':
        if (p.answer === 'ok') { transitionToChats(); send({ type: 'getChats' }); }
        else showError(t('err_create') + p.answer);
        break;
      case 'newChat':
        send({ type: 'getChats' });
        break;
      case 'msgStatus':
        if (p.answer === 'ok' && p.msgId) {
          const myMsgs = messagesList.querySelectorAll('.msg-row.own.pending:not([data-msg-id])');
          if (myMsgs.length > 0) {
            const lastMsg = myMsgs[myMsgs.length - 1];
            lastMsg.dataset.msgId = p.msgId;
            lastMsg.classList.remove('pending');
            updateChatPreview(p.chatId, lastMsg.dataset.originalText, myName, false, new Date().toISOString());
          }
        } else if (p.answer === 'failed') revertLastMessage(t('msg_server_err'));
        else if (p.answer === 'invalidMsgLength') revertLastMessage(t('msg_too_long'));
        else if (p.answer === 'slow_down') revertLastMessage(t('msg_slow_down'));
        break;
      case 'delMsg': {
        const el = messagesList.querySelector(`li.msg-row[data-msg-id="${p.msgId}"]`);
        if (el) el.remove();
        break;
      }
      case 'delMsgStatus':
        if (p.answer === 'failed') showError(t('err_del_msg'));
        break;
      case 'chatReadStatus': break;
      case 'chatRead': {
        if (p.chatId === currentChatId) {
          messagesList.querySelectorAll('.msg-row.own .msg-time .ticks').forEach(el => {
            el.classList.add('read'); el.textContent = '✔✔';
          });
        }
        const chatLi2 = chatList.querySelector(`li[data-chat-id="${p.chatId}"]`);
        if (chatLi2) {
          const ticks = chatLi2.querySelector('.chat-meta .ticks');
          if (ticks) { ticks.classList.add('read'); ticks.textContent = '✔✔'; }
        }
        break;
      }
      case 'delChatStatus':
        if (p.answer === 'ok') removeChatFromUI(p.chatId);
        else showError(t('err_del_chat'));
        break;
      case 'delChat': removeChatFromUI(p.chatId); break;
      case 'editMsgStatus': {
        if (p.answer === 'ok') {
          if (pendingEditMsgId !== null && pendingEditNewText !== null) {
            const editedLi = applyEditToDOM(pendingEditMsgId, pendingEditNewText);
            if (editedLi) {
              const chat = chatsById[pendingEditChatId];
              if (chat && chat.lastMsg === pendingEditOldText && chat.lastMsgSender === myName) {
                updateChatPreview(pendingEditChatId, pendingEditNewText, myName, chat.lastMsgIsRead, chat.lastUsed);
              }
            }
          }
          clearPendingEdit();
          editModal.style.display = 'none';
        } else if (p.answer === 'failed') showError(t('error'));
        break;
      }
      case 'editMsg': {
        const editedLi = applyEditToDOM(p.msgId, p.text);
        if (editedLi) {
          const chat = chatsById[currentChatId];
          if (chat) {
            let isLast = true;
            let next = editedLi.nextElementSibling;
            while (next) {
              if (next.classList.contains('msg-row')) { isLast = false; break; }
              next = next.nextElementSibling;
            }
            if (isLast) updateChatPreview(currentChatId, p.text, editedLi.dataset.sender, chat.lastMsgIsRead, chat.lastUsed);
          }
        }
        break;
      }
    }
  };
}

connectWebSocket();

document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible') {
    if (!ws || ws.readyState === WebSocket.CLOSED || ws.readyState === WebSocket.CLOSING) {
      clearTimeout(reconnectTimeout);
      connectWebSocket();
    }
  }
});

function send(obj) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  ws.send(JSON.stringify(obj));
}

function executeSend(msg) {
  if (msg.isGroup) send({ type: 'sendMsg', chatId: msg.chatId, text: msg.text, isGroup: true });
  else send({ type: 'sendMsg', to: msg.to, text: msg.text, isGroup: false });
}

function flushOutbox() {
  if (outbox.length === 0) return;
  const msgsToSend = [...outbox];
  outbox = [];
  msgsToSend.forEach(msg => {
    executeSend(msg);
    if (msg.chatId === currentChatId) {
      const li = addMessage(myName, msg.text, null, msg.timeStr);
      li.classList.add('pending');
    }
  });
}

// ================== DOM ELEMENTS ==================
const loginScreen    = document.getElementById('login-screen');
const regScreen      = document.getElementById('reg-screen');
const chatsScreen    = document.getElementById('chats');
const settingsScreen = document.getElementById('settings-screen');
const soonScreen     = document.getElementById('coming-soon');
const newChoiceScreen = document.getElementById('new-choice-screen');
const newChatScreen   = document.getElementById('new-chat-screen');
const newGroupScreen  = document.getElementById('new-group-screen');

const loginNameInput = document.getElementById('login-name');
const loginPassInput = document.getElementById('login-pass');
const regNameInput   = document.getElementById('reg-name');
const regPassInput   = document.getElementById('reg-pass');
const regPass2Input  = document.getElementById('reg-pass2');
const loginBtn       = document.getElementById('login-btn');
const registerBtn    = document.getElementById('reg-btn');
const toRegBtn       = document.getElementById('to-register');
const toLoginBtn     = document.getElementById('to-login');

const settingsBtn        = document.getElementById('settings-btn');
const backSettingsButton = document.getElementById('backSettingsButton');
const settingsProfileBtn = document.getElementById('settingsProfileBtn');
const settingsLogoutBtn  = document.getElementById('settingsLogoutBtn');
const settingsLangBtn    = document.getElementById('settingsLangBtn');
const settingsNotifBtn   = document.getElementById('settingsNotifBtn');
const soonBackBtn        = document.getElementById('soonBackBtn');

const chatList     = document.querySelector('.chat-list');
const messagesList = document.querySelector('.messages');
const sendForm     = document.querySelector('.send');
const sendInput    = sendForm.querySelector('input');
const chatTitle    = document.getElementById('chat-title');
const headerAvatar = document.getElementById('header-avatar');
const chatHeaderInfo = document.getElementById('chat-header-info');
const mobileBackBtn = document.getElementById('mobile-back-btn');

const newChatBtn      = document.getElementById('new-chat-btn');
const newChatInput    = document.getElementById('new-chat-input');
const createChatBtn   = document.getElementById('create-chat-btn');
const cancelChatBtn   = document.getElementById('cancel-chat-btn');

const choiceDmBtn     = document.getElementById('choice-dm-btn');
const choiceGroupBtn  = document.getElementById('choice-group-btn');
const choiceBackBtn   = document.getElementById('choice-back-btn');

const newGroupName        = document.getElementById('new-group-name');
const newGroupMemberInput = document.getElementById('new-group-member-input');
const groupChipsContainer = document.getElementById('group-chips-container');
const createGroupBtn      = document.getElementById('create-group-btn');
const cancelGroupBtn      = document.getElementById('cancel-group-btn');

const msgActionsBtn = document.getElementById('msg-actions-btn');
const msgPopover    = document.getElementById('msg-popover');
const dotsMenu      = document.getElementById('dots-menu');

const chatContextModal = document.getElementById('chat-context-modal');
const closeChatModalBtn = document.getElementById('close-chat-modal-btn');
const chatModalActionBtn = document.getElementById('chat-modal-action-btn');

const profileModal = document.getElementById('profile-modal');
const closeProfileModalBtn = document.getElementById('close-profile-modal-btn');
const groupMembersModal = document.getElementById('group-members-modal');
const closeGroupMembersModalBtn = document.getElementById('close-group-members-modal-btn');

const renameModal = document.getElementById('rename-modal');
const renameInput = document.getElementById('rename-input');
const saveRenameBtn = document.getElementById('save-rename-btn');
const closeRenameModalBtn = document.getElementById('close-rename-modal-btn');

const editModal = document.getElementById('edit-modal');
const editInput = document.getElementById('edit-input');
const saveEditBtn = document.getElementById('save-edit-btn');
const closeEditModalBtn = document.getElementById('close-edit-modal-btn');

const errorModal = document.getElementById('error-modal');
const errorText = document.getElementById('error-text');
const okErrorBtn = document.getElementById('ok-error-btn');
const closeErrorModalBtn = document.getElementById('close-error-modal-btn');

let renameTargetId = null;

function showError(msg) {
  errorText.textContent = msg;
  errorModal.style.display = 'flex';
}

closeErrorModalBtn.onclick = () => errorModal.style.display = 'none';
okErrorBtn.onclick = () => errorModal.style.display = 'none';
errorModal.onclick = (e) => { if (e.target === errorModal) errorModal.style.display = 'none'; };

// ================== SCREEN NAVIGATION ==================
function hideAllScreens() {
  [loginScreen, regScreen, chatsScreen, settingsScreen, soonScreen,
   newChoiceScreen, newChatScreen, newGroupScreen].forEach(el => el.style.display = 'none');
}

function showLogin() {
  hideAllScreens(); loginScreen.style.display = 'flex';
  history.replaceState({ view: 'login' }, '');
}
function showReg() {
  hideAllScreens(); regScreen.style.display = 'flex';
  history.replaceState({ view: 'reg' }, '');
}
function showSettings() { navigate('settings'); }
function showComingSoon() { navigate('coming-soon'); }
function showNewChoice() { navigate('new-choice'); }
function showNewChat() { navigate('new-chat'); }
function showNewGroup() { navigate('new-group'); }

toRegBtn.onclick = showReg;
toLoginBtn.onclick = showLogin;
soonBackBtn.onclick = () => history.back();
backSettingsButton.onclick = () => history.back();
choiceBackBtn.onclick = () => history.back();
cancelChatBtn.onclick = () => history.back();
cancelGroupBtn.onclick = () => history.back();
mobileBackBtn.onclick = (e) => { e.stopPropagation(); closeActiveChat(); };

chatHeaderInfo.onclick = () => {
  if (!currentChatId) return;
  const chat = chatsById[currentChatId];
  if (!chat) return;
  if (chat.isGroup) {
    document.getElementById('group-members-title').textContent = chat.chatName;
    groupMembersModal.style.display = 'flex';
    send({ type: 'getGroupMembers', chatId: currentChatId });
  } else {
    document.getElementById('profile-name').textContent = chat.chatName;
    document.getElementById('profile-avatar').textContent = chat.chatName.charAt(0).toUpperCase();
    profileModal.style.display = 'flex';
  }
};

closeProfileModalBtn.onclick = () => profileModal.style.display = 'none';
profileModal.onclick = (e) => { if (e.target === profileModal) profileModal.style.display = 'none'; };
closeGroupMembersModalBtn.onclick = () => groupMembersModal.style.display = 'none';
groupMembersModal.onclick = (e) => { if (e.target === groupMembersModal) groupMembersModal.style.display = 'none'; };

// === CUSTOM NAME ===
document.getElementById('rename-profile-btn').onclick = () => {
  renameTargetId = currentChatId;
  renameInput.value = document.getElementById('profile-name').textContent;
  profileModal.style.display = 'none';
  renameModal.style.display = 'flex';
  renameInput.focus();
};

document.getElementById('reset-profile-btn').onclick = () => {
  send({ type: 'delCustomName', chatId: currentChatId });
};

closeRenameModalBtn.onclick = () => renameModal.style.display = 'none';
renameModal.onclick = (e) => { if (e.target === renameModal) renameModal.style.display = 'none'; };

saveRenameBtn.onclick = () => {
  const newName = renameInput.value.trim();
  if (newName && renameTargetId) {
    pendingRenameId = renameTargetId;
    pendingRenameName = newName;
    send({ type: 'makeCustomName', chatId: renameTargetId, name: newName });
  } else renameModal.style.display = 'none';
};

renameInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') saveRenameBtn.click(); });

// ================== EDIT MESSAGE ==================
function startEditMessage(msgId) {
  const li = messagesList.querySelector(`li.msg-row[data-msg-id="${msgId}"]`);
  if (!li) return;
  pendingEditMsgId = parseInt(msgId);
  pendingEditChatId = currentChatId;
  pendingEditOldText = li.dataset.originalText;
  editInput.value = pendingEditOldText;
  editModal.style.display = 'flex';
  editInput.focus();
  try { editInput.setSelectionRange(editInput.value.length, editInput.value.length); } catch {}
}

function clearPendingEdit() {
  pendingEditMsgId = null; pendingEditChatId = null;
  pendingEditOldText = null; pendingEditNewText = null;
}

closeEditModalBtn.onclick = () => { editModal.style.display = 'none'; clearPendingEdit(); };
editModal.onclick = (e) => { if (e.target === editModal) { editModal.style.display = 'none'; clearPendingEdit(); } };

saveEditBtn.onclick = () => {
  const newText = editInput.value.trim();
  if (!newText) return;
  if (!pendingEditMsgId || pendingEditChatId === null) { clearPendingEdit(); editModal.style.display = 'none'; return; }
  if (newText === pendingEditOldText) { editModal.style.display = 'none'; clearPendingEdit(); return; }
  if (newText.length > 1000) { showError(t('msg_too_long')); return; }
  pendingEditNewText = newText;
  send({ type: 'editMsg', chatId: pendingEditChatId, msgId: pendingEditMsgId, text: newText });
};

editInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') saveEditBtn.click(); });

function applyEditToDOM(msgId, newText) {
  const li = messagesList.querySelector(`li.msg-row[data-msg-id="${msgId}"]`);
  if (!li) return null;
  const textSpan = li.querySelector('.msg-text');
  const sender = li.dataset.sender;
  const chat = chatsById[currentChatId];
  const isGroup = chat ? chat.isGroup : false;
  const showName = isGroup && sender !== myName;
  const displayName = getDisplayName(sender);

  if (textSpan) textSpan.textContent = showName ? (displayName + ': ' + newText) : newText;
  li.dataset.originalText = newText;

  const timeEl = li.querySelector('.msg-time');
  if (timeEl && !timeEl.querySelector('.edited-mark')) {
    const mark = document.createElement('span');
    mark.className = 'edited-mark';
    mark.textContent = '(edited)';
    const ticks = timeEl.querySelector('.ticks');
    if (ticks) timeEl.insertBefore(mark, ticks); else timeEl.appendChild(mark);
  }
  return li;
}

// ================== AUTHENTICATION ==================
loginBtn.onclick = () => {
  const name = loginNameInput.value.trim();
  if (!name) return showError(t('name') + ' required');
  send({ type: 'login', name, password: loginPassInput.value });
};

registerBtn.onclick = () => {
  const name  = regNameInput.value.trim();
  const pass  = regPassInput.value;
  const pass2 = regPass2Input.value;
  if (!/^[a-zA-Z0-9_а-яА-ЯёЁ]{3,20}$/.test(name)) return showError('Name must be 3-20 characters');
  if (pass !== pass2) return showError(t('password') + ' mismatch');
  send({ type: 'reg', name, password: pass });
};

[loginNameInput, loginPassInput].forEach(el => el.addEventListener('keydown', e => { if (e.key === 'Enter') loginBtn.click(); }));
[regNameInput, regPassInput, regPass2Input].forEach(el => el.addEventListener('keydown', e => { if (e.key === 'Enter') registerBtn.click(); }));

settingsBtn.onclick = showSettings;
settingsProfileBtn.onclick = showComingSoon;
settingsLangBtn.onclick = () => {
  currentLang = currentLang === 'en' ? 'ru' : 'en';
  localStorage.setItem('lang', currentLang);
  applyTranslations();
  renderChatList(Object.values(chatsById));
};
settingsLogoutBtn.onclick = () => { localStorage.removeItem('token'); location.reload(); };

// ================== CHAT CREATION ==================
newChatBtn.onclick = showNewChoice;
choiceDmBtn.onclick = () => { newChatInput.value = ''; showNewChat(); };
choiceGroupBtn.onclick = () => { newGroupName.value = ''; newGroupMemberInput.value = ''; groupMembers = []; renderChips(); showNewGroup(); };

function renderChips() {
  groupChipsContainer.innerHTML = '';
  groupMembers.forEach((name, index) => {
    const chip = document.createElement('div');
    chip.className = 'group-chip';
    chip.textContent = name;
    const btn = document.createElement('button');
    btn.textContent = '×';
    btn.onclick = () => { groupMembers.splice(index, 1); renderChips(); };
    chip.appendChild(btn);
    groupChipsContainer.appendChild(chip);
  });
}

function cleanMemberName(str) { return str.replace(/[\s,]+/g, ''); }

function addMemberFromInput() {
  const name = cleanMemberName(newGroupMemberInput.value);
  if (name && !groupMembers.includes(name)) { groupMembers.push(name); renderChips(); }
  newGroupMemberInput.value = '';
}

newGroupMemberInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' || e.key === ',') { e.preventDefault(); addMemberFromInput(); }
});

createChatBtn.onclick = () => {
  const name = newChatInput.value.trim();
  if (name) send({ type: 'createChat', to: name, isGroup: false });
};

createGroupBtn.onclick = () => {
  addMemberFromInput();
  const gName = newGroupName.value.trim();
  if (!gName) return showError(t('group_name') + ' required');
  if (groupMembers.length === 0) return showError(t('add_member'));
  send({ type: 'createChat', isGroup: true, name: gName, data: groupMembers });
};

// ================== GROUP MEMBERS ==================
function renderGroupMembers(members) {
  const list = document.getElementById('group-members-list');
  list.innerHTML = '';
  if (!members || members.length === 0) {
    list.innerHTML = `<li style="justify-content:center; border:none;">${t('no_members')}</li>`;
    return;
  }
  members.forEach(m => {
    const li = document.createElement('li');
    const avatar = document.createElement('div');
    avatar.className = 'chat-avatar';
    const displayName = getDisplayName(m);
    avatar.textContent = displayName.charAt(0).toUpperCase();
    const name = document.createElement('span');
    name.textContent = displayName;
    li.appendChild(avatar); li.appendChild(name); list.appendChild(li);
  });
}

// ================== CHAT LIST ==================
function moveChatToTop(chatId) {
  const li = chatList.querySelector(`li[data-chat-id="${chatId}"]`);
  if (li && chatList.firstChild !== li) chatList.prepend(li);
}

function clearUnreadUI(chatId) {
  const li = chatList.querySelector(`li[data-chat-id="${chatId}"]`);
  if (!li) return;
  li.classList.remove('unread');
  const badge = li.querySelector('.unread-badge');
  if (badge) badge.remove();
}

function setUnreadUI(chatId, count) {
  const li = chatList.querySelector(`li[data-chat-id="${chatId}"]`);
  if (!li) return;
  let badge = li.querySelector('.unread-badge');
  if (count > 0) {
    li.classList.add('unread');
    if (!badge) {
      badge = document.createElement('span');
      badge.className = 'unread-badge';
      const bottomDiv = li.querySelector('.chat-bottom');
      if (bottomDiv) bottomDiv.appendChild(badge);
    }
    badge.textContent = count > 99 ? '99+' : count;
  } else clearUnreadUI(chatId);
}

function updateChatPreview(chatId, text, sender, isRead, timeStr) {
  const li = chatList.querySelector(`li[data-chat-id="${chatId}"]`);
  if (!li) return;
  const chat = chatsById[chatId];
  if (!chat) return;

  const bottomDiv = li.querySelector('.chat-bottom');
  const metaSpan = li.querySelector('.chat-meta');

  if (metaSpan) {
    let ticksHTML = '';
    if (sender === myName) ticksHTML = `<span class="ticks ${isRead ? 'read' : ''}">${isRead ? '✔✔' : '✔'}</span>`;
    metaSpan.innerHTML = `${ticksHTML}<span>${formatMsgTime(timeStr)}</span>`;
  }

  const msgSpan = bottomDiv.querySelector('.chat-last-msg');
  if (msgSpan) {
    const displaySender = getDisplayName(sender);
    if (chat.isGroup && sender && sender !== 'SYSTEM' && sender !== myName) msgSpan.textContent = displaySender + ': ' + text;
    else msgSpan.textContent = text;
  }

  chat.lastMsg = text; chat.lastMsgSender = sender; chat.lastMsgIsRead = isRead; chat.lastUsed = timeStr;
  moveChatToTop(chatId);
}

function renderChatList(chats) {
  chatList.innerHTML = '';
  Object.keys(chatsById).forEach(k => delete chatsById[k]);

  if (!chats.length) {
    chatTitle.textContent = '';
    messagesList.innerHTML = '';
    return;
  }

  chats.forEach(c => {
    chatsById[c.id] = c;
    const li = document.createElement('li');
    li.dataset.chatId = c.id; li.dataset.chatName = c.chatName;

    const avatar = document.createElement('div');
    avatar.className = 'chat-avatar';
    avatar.textContent = c.chatName.charAt(0).toUpperCase();
    li.appendChild(avatar);

    const infoDiv = document.createElement('div'); infoDiv.className = 'chat-info';
    const topDiv = document.createElement('div'); topDiv.className = 'chat-top';

    const nameSpan = document.createElement('span'); nameSpan.className = 'chat-name';
    nameSpan.textContent = c.chatName; topDiv.appendChild(nameSpan);

    const metaSpan = document.createElement('span'); metaSpan.className = 'chat-meta';
    if (c.lastMsgSender === myName) {
      const ticksEl = document.createElement('span');
      ticksEl.className = c.lastMsgIsRead ? 'ticks read' : 'ticks';
      ticksEl.textContent = c.lastMsgIsRead ? '✔✔' : '✔';
      metaSpan.appendChild(ticksEl);
    }
    const timeEl = document.createElement('span');
    timeEl.textContent = formatMsgTime(c.lastUsed);
    metaSpan.appendChild(timeEl); topDiv.appendChild(metaSpan); infoDiv.appendChild(topDiv);

    const bottomDiv = document.createElement('div'); bottomDiv.className = 'chat-bottom';
    const msgSpan = document.createElement('span'); msgSpan.className = 'chat-last-msg';
    const displaySender = getDisplayName(c.lastMsgSender);
    if (c.isGroup && c.lastMsgSender && c.lastMsgSender !== 'SYSTEM' && c.lastMsgSender !== myName) {
      msgSpan.textContent = displaySender + ': ' + c.lastMsg;
    } else msgSpan.textContent = c.lastMsg;
    bottomDiv.appendChild(msgSpan);

    if (c.unread > 0) {
      const badge = document.createElement('span'); badge.className = 'unread-badge';
      badge.textContent = c.unread > 99 ? '99+' : c.unread; bottomDiv.appendChild(badge);
    }

    infoDiv.appendChild(bottomDiv); li.appendChild(infoDiv); chatList.appendChild(li);

    if (c.unread > 0) li.classList.add('unread');
    if (c.id === currentChatId) li.classList.add('active');
  });

  if (currentChatId && chatsById[currentChatId]) {
    chatTitle.textContent = chatsById[currentChatId].chatName;
    headerAvatar.textContent = chatsById[currentChatId].chatName.charAt(0).toUpperCase();
  }
}

chatList.onclick = (e) => {
  const li = e.target.closest('li');
  if (li) {
    const chatId = parseInt(li.dataset.chatId);
    if (currentChatId !== chatId) navigate('chat', { chatId });
  }
};

chatList.addEventListener('contextmenu', (e) => {
  const li = e.target.closest('li'); if (!li) return;
  e.preventDefault(); openChatContextMenu(li);
});

chatList.addEventListener('touchstart', (e) => {
  const li = e.target.closest('li'); if (!li) return;
  chatTouchTimer = setTimeout(() => { openChatContextMenu(li); if (navigator.vibrate) navigator.vibrate(50); }, 500);
}, { passive: true });

chatList.addEventListener('touchend', () => clearTimeout(chatTouchTimer));
chatList.addEventListener('touchmove', () => clearTimeout(chatTouchTimer));
chatList.addEventListener('touchcancel', () => clearTimeout(chatTouchTimer));

function openChatContextMenu(li) {
  const chatId = parseInt(li.dataset.chatId); const chat = chatsById[chatId];
  if (!chat) return;
  selectedChatIdForAction = chatId;
  chatModalActionBtn.textContent = chat.isGroup ? t('leave_group') : t('delete_chat');
  chatContextModal.style.display = 'flex';
}

chatModalActionBtn.onclick = () => {
  if (selectedChatIdForAction) { send({ type: 'delChat', chatId: selectedChatIdForAction }); chatContextModal.style.display = 'none'; }
};

closeChatModalBtn.onclick = () => { chatContextModal.style.display = 'none'; };
chatContextModal.onclick = (e) => { if (e.target === chatContextModal) chatContextModal.style.display = 'none'; };

function removeChatFromUI(chatId) {
  const li = chatList.querySelector(`li[data-chat-id="${chatId}"]`);
  if (li) li.remove();
  if (currentChatId === chatId) closeActiveChat();
  delete chatsById[chatId];
}

function closeActiveChat() {
  if (chatContextModal.style.display === 'flex') { chatContextModal.style.display = 'none'; return; }
  closeMsgPopover(); closeDotsMenu();
  if (currentChatId) { history.replaceState({ view: 'chats' }, ''); renderView({ view: 'chats' }); }
}

// ================== MESSAGE RENDERING ==================
messagesList.onscroll = () => {
  if (msgPopover.classList.contains('open')) closeMsgPopover();
  if (dotsMenu.classList.contains('open')) closeDotsMenu();
  if (messagesList.scrollTop === 0 && hasMoreHistory && !isLoadingHistory && currentChatId && oldestMessageId) {
    isLoadingHistory = true;
    send({ type: 'getHistory', chatId: currentChatId, beforeId: oldestMessageId });
  }
};

function createMessageEl(sender, text, msgId, timeStr, isRead, isEdited) {
  const li = document.createElement('li');
  if (sender === 'SYSTEM') {
    li.className = 'system-msg';
    let sysText = text;
    if (currentLang === 'ru' && sysText.endsWith(" left the group")) {
      sysText = sysText.replace(" left the group", ` ${t('left_group')}`);
    }
    li.textContent = sysText;
    if (msgId) li.dataset.msgId = msgId;
    return li;
  }

  li.className = 'msg-row ' + ((sender === myName) ? 'own' : 'other');
  li.dataset.originalText = text;
  li.dataset.time = timeStr || new Date().toISOString();
  li.dataset.sender = sender;

  const bubble = document.createElement('div'); bubble.className = 'msg-bubble';
  const textSpan = document.createElement('span'); textSpan.className = 'msg-text';

  const chat = chatsById[currentChatId];
  const isGroup = chat ? chat.isGroup : false;
  const displaySender = getDisplayName(sender);
  const showName = isGroup && (sender !== myName);
  textSpan.textContent = showName ? (displaySender + ': ' + text) : text;
  bubble.appendChild(textSpan);

  const timeEl = document.createElement('span'); timeEl.className = 'msg-time';
  timeEl.appendChild(document.createTextNode(formatMsgTime(timeStr)));

  if (isEdited) {
    const editedMark = document.createElement('span'); editedMark.className = 'edited-mark';
    editedMark.textContent = '(edited)'; timeEl.appendChild(editedMark);
  }

  if (sender === myName) {
    const ticksEl = document.createElement('span');
    ticksEl.className = isRead ? 'ticks read' : 'ticks';
    ticksEl.textContent = isRead ? '✔✔' : '✔';
    timeEl.appendChild(ticksEl);
  }

  bubble.appendChild(timeEl); li.appendChild(bubble);
  if (msgId) li.dataset.msgId = msgId;
  if (sender === myName && currentChatId) li.dataset.chatId = currentChatId;
  return li;
}

function renderHistory(msgs) {
  if (msgs.length === 0) { hasMoreHistory = false; return; }

  if (oldestMessageId !== null) {
    const prevHeight = messagesList.scrollHeight;
    const fragment = document.createDocumentFragment();
    let currentBatchDate = null;
    for (let i = 0; i < msgs.length; i++) {
      const m = msgs[i];
      const dateStr = getDateStr(m.time);
      if (dateStr !== currentBatchDate) {
        const divider = document.createElement('li'); divider.className = 'date-divider';
        divider.textContent = formatDateDivider(m.time); fragment.appendChild(divider); currentBatchDate = dateStr;
      }
      fragment.appendChild(createMessageEl(m.sender, m.message, m.id, m.time, m.isRead, m.isEdited));
    }
    messagesList.insertBefore(fragment, messagesList.firstChild);
    messagesList.scrollTop += (messagesList.scrollHeight - prevHeight);
  } else {
    messagesList.innerHTML = '';
    let currentBatchDate = null;
    for (let i = 0; i < msgs.length; i++) {
      const m = msgs[i];
      const dateStr = getDateStr(m.time);
      if (dateStr !== currentBatchDate) {
        const divider = document.createElement('li'); divider.className = 'date-divider';
        divider.textContent = formatDateDivider(m.time); messagesList.appendChild(divider); currentBatchDate = dateStr;
      }
      messagesList.appendChild(createMessageEl(m.sender, m.message, m.id, m.time, m.isRead, m.isEdited));
    }
    messagesList.scrollTop = messagesList.scrollHeight;
    if (msgs.length > 0) lastAppendedDate = getDateStr(msgs[msgs.length - 1].time);
  }
  oldestMessageId = msgs[0].id;
}

function addMessage(sender, text, msgId, timeStr, isEdited = false) {
  const isAtBottom = messagesList.scrollHeight - messagesList.scrollTop - messagesList.clientHeight < 50;
  const msgDate = getDateStr(timeStr);
  if (lastAppendedDate && msgDate !== lastAppendedDate) {
    const divider = document.createElement('li'); divider.className = 'date-divider';
    divider.textContent = formatDateDivider(timeStr); messagesList.appendChild(divider);
  }
  const li = createMessageEl(sender, text, msgId, timeStr, false, isEdited);
  messagesList.appendChild(li);
  if (sender === myName || isAtBottom) messagesList.scrollTop = messagesList.scrollHeight;
  lastAppendedDate = msgDate;
  return li;
}

// ================== MESSAGE SENDING ==================
sendForm.onsubmit = (e) => {
  e.preventDefault();
  const text = sendInput.value.trim();
  if (!text || !currentChatId) return;
  const chat = chatsById[currentChatId];
  if (!chat) return;
  closeMsgPopover();

  const timeStr = new Date().toISOString();
  const msgObj = { chatId: currentChatId, isGroup: chat.isGroup, to: chat.peerName || chat.chatName, text: text, timeStr: timeStr };

  if (ws && ws.readyState === WebSocket.OPEN) {
    executeSend(msgObj);
    const li = addMessage(myName, text, null, timeStr);
    li.classList.add('pending');
    updateChatPreview(currentChatId, text, myName, false, timeStr);
  } else {
    outbox.push(msgObj);
    const li = addMessage(myName, text, null, timeStr);
    li.classList.add('pending');
    updateChatPreview(currentChatId, text, myName, false, timeStr);
  }
  sendInput.value = '';
  sendInput.focus();
};

function revertLastMessage(errorMsg) {
  showError(errorMsg);
  const myMsgs = messagesList.querySelectorAll('.msg-row.own.pending:not([data-msg-id])');
  if (myMsgs.length > 0) {
    const lastMsg = myMsgs[myMsgs.length - 1];
    const originalText = lastMsg.dataset.originalText;
    if (sendInput.value.trim() === '') { sendInput.value = originalText; sendInput.focus(); }
    lastMsg.remove();
  }
}

function handleIncomingMsg(p) {
  const msgId = p.msgId || p.id;
  const timeStr = p.time || new Date().toISOString();
  let targetChatId = p.chatId;
  if (!targetChatId) {
    const foundChat = Object.values(chatsById).find(c => !c.isGroup && c.peerName === p.from);
    if (foundChat) targetChatId = foundChat.id;
  }
  if (!targetChatId) return;

  if (targetChatId === currentChatId) {
    addMessage(p.from, p.text, msgId, timeStr);
    send({ type: 'chatRead', chatId: currentChatId });
    updateChatPreview(targetChatId, p.text, p.from, false, timeStr);
  } else {
    const li = chatList.querySelector(`li[data-chat-id="${targetChatId}"]`);
    if (li) {
      const badge = li.querySelector('.unread-badge');
      let count = badge ? parseInt(badge.textContent) : 0;
      if (count === 99) count = 100; else if (count < 99) count++;
      setUnreadUI(targetChatId, count);
      updateChatPreview(targetChatId, p.text, p.from, false, timeStr);
    }
    if (document.hidden && Notification.permission === 'granted') {
      const notification = new Notification(getDisplayName(p.from), { body: p.text, icon: '/favicon.png' });
      notification.onclick = () => { window.focus(); notification.close(); };
    }
  }
}

// ================== SELECTION & POPOVERS ==================
const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0;

messagesList.addEventListener('mousedown', (e) => {
  if (e.button !== 0) return;
  if (msgPopover.classList.contains('open')) return;
  const li = e.target.closest('li.msg-row');
  if (!li || !li.dataset.msgId) return;

  mouseDownTarget = li; dragStartX = e.clientX; dragStartY = e.clientY;
  isDragging = false; wasDragging = false;
  dragRows = Array.from(messagesList.querySelectorAll('li.msg-row'));
});

document.addEventListener('mousemove', (e) => {
  if (!mouseDownTarget) return;
  if (!isDragging && (Math.abs(e.clientX - dragStartX) > 5 || Math.abs(e.clientY - dragStartY) > 5)) {
    isDragging = true; document.body.style.userSelect = 'none';
  }
  if (isDragging) {
    const el = document.elementFromPoint(e.clientX, e.clientY);
    const li = el?.closest('li.msg-row');
    if (li && dragRows.length > 0) selectRange(mouseDownTarget, li, dragRows);
  }
});

document.addEventListener('mouseup', (e) => {
  if (e.button !== 0) return;
  if (isDragging) wasDragging = true;
  mouseDownTarget = null; isDragging = false; document.body.style.userSelect = '';
});

function selectRange(startLi, endLi, allRows) {
  const startIndex = allRows.indexOf(startLi); const endIndex = allRows.indexOf(endLi);
  if (startIndex === -1 || endIndex === -1) return;
  const min = Math.min(startIndex, endIndex); const max = Math.max(startIndex, endIndex);
  selectedMsgIds.clear();
  document.querySelectorAll('.messages li.msg-row.selected').forEach(el => el.classList.remove('selected'));
  for (let i = min; i <= max; i++) {
    const li = allRows[i];
    if (li.dataset.msgId) { selectedMsgIds.add(li.dataset.msgId); li.classList.add('selected'); }
  }
  if (selectedMsgIds.size > 0) { msgActionsBtn.style.display = 'block'; msgActionsBtn.textContent = `⋮ (${selectedMsgIds.size})`; }
  else msgActionsBtn.style.display = 'none';
}

messagesList.addEventListener('contextmenu', (e) => {
  const li = e.target.closest('li.msg-row');
  if (!li || !li.dataset.msgId) return;
  e.preventDefault();
  if (isTouchDevice) return;
  if (msgPopover.classList.contains('open')) { closeMsgPopover(); return; }
  if (selectedMsgIds.size > 0) { clearSelection(); closeDotsMenu(); }
  openMsgPopover(li, e.clientX, e.clientY);
});

messagesList.addEventListener('touchstart', (e) => {
  const li = e.target.closest('li.msg-row');
  if (!li || !li.dataset.msgId) return;
  isLongPress = false;
  touchTimer = setTimeout(() => {
    isLongPress = true; toggleSelection(li);
    if (navigator.vibrate) navigator.vibrate(50);
  }, 500);
}, { passive: true });

messagesList.addEventListener('touchmove', () => clearTimeout(touchTimer), { passive: true });
messagesList.addEventListener('touchend', () => clearTimeout(touchTimer));
messagesList.addEventListener('touchcancel', () => clearTimeout(touchTimer));

messagesList.addEventListener('click', (e) => {
  if (wasDragging) { wasDragging = false; return; }
  if (msgPopover.classList.contains('open')) return;
  const li = e.target.closest('li.msg-row');
  if (selectedMsgIds.size > 0) {
    if (isTouchDevice && li && li.dataset.msgId) toggleSelection(li);
    else { clearSelection(); closeDotsMenu(); }
    return;
  }
  if (!isTouchDevice) return;
  if (isLongPress) { isLongPress = false; return; }
  if (!li || !li.dataset.msgId) return;
  e.stopPropagation();
  if (openPopoverMsgId === li.dataset.msgId) closeMsgPopover();
  else openMsgPopover(li, e.clientX, e.clientY);
});

function toggleSelection(li) {
  const id = li.dataset.msgId;
  if (selectedMsgIds.has(id)) { selectedMsgIds.delete(id); li.classList.remove('selected'); }
  else { selectedMsgIds.add(id); li.classList.add('selected'); }
  if (selectedMsgIds.size > 0) { msgActionsBtn.style.display = 'block'; msgActionsBtn.textContent = `⋮ (${selectedMsgIds.size})`; }
  else msgActionsBtn.style.display = 'none';
}

function clearSelection() {
  selectedMsgIds.clear();
  document.querySelectorAll('.messages li.msg-row.selected').forEach(el => el.classList.remove('selected'));
  msgActionsBtn.style.display = 'none'; closeDotsMenu();
}

function positionPopover(popover, x, y) {
  popover.style.left = '0px'; popover.style.top = '0px';
  const rect = popover.getBoundingClientRect(); const w = rect.width, h = rect.height;
  const vw = window.innerWidth, vh = window.innerHeight;
  let left = x, top = y;
  if (left + w > vw - 8) left = vw - w - 8;
  if (left < 8) left = 8;
  if (top + h > vh - 8) top = vh - h - 8;
  if (top < 8) top = 8;
  popover.style.left = left + 'px'; popover.style.top = top + 'px';
}

function openMsgPopover(li, x, y) {
  closeDotsMenu();
  const msgId = li.dataset.msgId; if (!msgId) return;
  openPopoverMsgId = msgId;
  const isOwn = li.dataset.sender === myName;
  msgPopover.innerHTML = '';

  if (isOwn) {
    const editBtn = document.createElement('button');
    editBtn.textContent = t('edit');
    editBtn.onclick = (ev) => { ev.stopPropagation(); startEditMessage(msgId); closeMsgPopover(); };
    msgPopover.appendChild(editBtn);
  }
  const copyBtn = document.createElement('button');
  copyBtn.textContent = t('copy');
  copyBtn.onclick = (ev) => {
    ev.stopPropagation(); const txt = li.dataset.originalText || '';
    if (navigator.clipboard) navigator.clipboard.writeText(txt).catch(() => {});
    closeMsgPopover();
  };
  msgPopover.appendChild(copyBtn);

  const forwardBtn = document.createElement('button');
  forwardBtn.textContent = t('forward');
  forwardBtn.onclick = (ev) => { ev.stopPropagation(); showToast(t('forward_soon')); closeMsgPopover(); };
  msgPopover.appendChild(forwardBtn);

  const deleteBtn = document.createElement('button');
  deleteBtn.textContent = t('delete');
  deleteBtn.onclick = (ev) => {
    ev.stopPropagation();
    send({ type: 'delMsg', msgId: parseInt(msgId), chatId: currentChatId });
    const el = messagesList.querySelector(`li.msg-row[data-msg-id="${msgId}"]`);
    if (el) el.remove();
    closeMsgPopover();
  };
  msgPopover.appendChild(deleteBtn);

  msgPopover.classList.add('open');
  positionPopover(msgPopover, x, y);
}

function closeMsgPopover() { openPopoverMsgId = null; msgPopover.classList.remove('open'); }

function openDotsMenu() {
  if (selectedMsgIds.size === 0) return;
  closeMsgPopover(); dotsMenu.innerHTML = '';

  const copyBtn = document.createElement('button');
  copyBtn.textContent = t('copy');
  copyBtn.onclick = (ev) => { ev.stopPropagation(); copySelectedMessages(); closeDotsMenu(); clearSelection(); };
  dotsMenu.appendChild(copyBtn);

  const forwardBtn = document.createElement('button');
  forwardBtn.textContent = t('forward');
  forwardBtn.onclick = (ev) => { ev.stopPropagation(); showToast(t('forward_soon')); closeDotsMenu(); };
  dotsMenu.appendChild(forwardBtn);

  const deleteBtn = document.createElement('button');
  deleteBtn.textContent = t('delete');
  deleteBtn.onclick = (ev) => { ev.stopPropagation(); deleteSelectedMessages(); closeDotsMenu(); clearSelection(); };
  dotsMenu.appendChild(deleteBtn);

  dotsMenu.classList.add('open');
  const btnRect = msgActionsBtn.getBoundingClientRect();
  positionPopover(dotsMenu, btnRect.left, btnRect.bottom + 8);
}

function closeDotsMenu() { dotsMenu.classList.remove('open'); }

function copySelectedMessages() {
  const textToCopy = [];
  const sortedIds = Array.from(selectedMsgIds).sort((a, b) => parseInt(a) - parseInt(b));
  sortedIds.forEach(id => {
    const el = messagesList.querySelector(`li.msg-row[data-msg-id="${id}"]`);
    if (el && el.dataset.originalText) textToCopy.push(el.dataset.originalText);
  });
  if (textToCopy.length > 0 && navigator.clipboard) {
    navigator.clipboard.writeText(textToCopy.join('\n\n')).catch(() => { showError(t('err_copy')); });
  }
}

function deleteSelectedMessages() {
  selectedMsgIds.forEach(id => {
    send({ type: 'delMsg', msgId: parseInt(id), chatId: currentChatId });
    const el = messagesList.querySelector(`li.msg-row[data-msg-id="${id}"]`);
    if (el) el.remove();
  });
}

msgActionsBtn.onclick = (e) => {
  e.stopPropagation();
  if (selectedMsgIds.size === 0) return;
  if (dotsMenu.classList.contains('open')) closeDotsMenu(); else openDotsMenu();
};

document.addEventListener('click', (e) => {
  if (msgPopover.classList.contains('open') && !msgPopover.contains(e.target)) closeMsgPopover();
  if (dotsMenu.classList.contains('open') && !dotsMenu.contains(e.target) && !e.target.closest('#msg-actions-btn')) closeDotsMenu();
});

document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    if (msgPopover.classList.contains('open')) closeMsgPopover();
    else if (dotsMenu.classList.contains('open')) closeDotsMenu();
    else if (chatContextModal.style.display === 'flex') chatContextModal.style.display = 'none';
    else if (profileModal.style.display === 'flex') profileModal.style.display = 'none';
    else if (groupMembersModal.style.display === 'flex') groupMembersModal.style.display = 'none';
    else if (renameModal.style.display === 'flex') renameModal.style.display = 'none';
    else if (editModal.style.display === 'flex') { editModal.style.display = 'none'; clearPendingEdit(); }
    else if (errorModal.style.display === 'flex') errorModal.style.display = 'none';
    else if (currentChatId) closeActiveChat();
    else if (history.state && history.state.view !== 'chats' && history.state.view !== 'login') history.back();
  }
});

// ================== PWA & PUSH ==================
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js').catch(err => console.log('SW error:', err));
}

if (window.visualViewport) {
  const fixKeyboardScroll = () => {
    const app = document.querySelector('.app');
    if (!app) return;
    app.style.height = `${window.visualViewport.height}px`;
    if (window.visualViewport.offsetTop > 0) window.scrollTo(0, 0);
  };
  window.visualViewport.addEventListener('resize', fixKeyboardScroll);
  window.visualViewport.addEventListener('scroll', fixKeyboardScroll);
}

const msgInput = document.querySelector('.send input');
if (msgInput) {
  msgInput.addEventListener('focus', () => { setTimeout(() => window.scrollTo(0, 0), 100); });
}

settingsNotifBtn.onclick = subscribeToPush;

function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding).replace(/\-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) outputArray[i] = rawData.charCodeAt(i);
  return outputArray;
}

async function subscribeToPush() {
  const permission = await Notification.requestPermission();
  if (permission !== 'granted') return showError(t('no_permission'));
  const res = await fetch('/api/vapid_public');
  const publicKey = await res.text();
  const reg = await navigator.serviceWorker.ready;
  const sub = await reg.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: urlBase64ToUint8Array(publicKey)
  });
  send({ type: 'savePushSub', data: sub });
  showToast(t('notifs_enabled'));
}