import { MapEngine } from './js/mapEngine.js';

let ws = null;
let connected = false;
let isLogin = false;

const MAX_HEALTH = 100;
const MAX_MANA = 50;
const MAX_STAMINA = 100;

let playerX = 512, playerY = 512;
let playerHealth = MAX_HEALTH, playerMana = MAX_MANA, playerStamina = MAX_STAMINA;
let maxHealth = MAX_HEALTH, maxMana = MAX_MANA, maxStamina = MAX_STAMINA;
let monsterHealth = 80;
let isInBattle = false;
let currentTarget = null;
let sessionID = '';
let playerID = 0;
let battleID = 0;
let playerLevel = 1;
let playerExp = 0;
let playerStrength = 10;
let playerAgility = 10;
let playerIntelligence = 10;
let playerDefense = 5;

let inventoryItems = {};
let equipments = {};
let playerGold = 1000;
let inventoryCapacity = 100;
let itemConfigs = {};

const ITEM_ICONS = {
    1001: '🧪', 1002: '🧪', 1003: '🧪',
    2001: '⚗️', 2002: '⚗️',
    3001: '📜', 3002: '📜',
    4001: '⚔️', 4002: '⚔️',
    4003: '🛡️', 4004: '⛑️',
    5001: '🪨', 5002: '🪵', 5003: '💎',
    6001: '🔑', 6002: '🏅',
    7001: '💰', 7002: '⭐'
};

const EQUIP_SLOT_NAMES = {
    1: '武器', 2: '护甲', 3: '头盔', 4: '靴子', 5: '手套', 6: '戒指', 7: '项链'
};

const SKILL_CONFIG = {
    1: {name: '普通攻击', damage: 15, mana: 0, effectType: 'normal', projectile: false},
    2: {name: '火球术', damage: 30, mana: 10, effectType: 'fire', projectile: true},
    3: {name: '冰霜箭', damage: 25, mana: 8, effectType: 'ice', projectile: true},
    4: {name: '闪电链', damage: 20, mana: 12, effectType: 'lightning', projectile: true},
    100: {name: '治疗术', damage: -30, mana: 15, effectType: 'heal', projectile: false},
    101: {name: '护盾', damage: 0, mana: 10, effectType: 'shield', projectile: false},
    102: {name: '加速', damage: 0, mana: 5, effectType: 'lightning', projectile: false},
    103: {name: '瞬移', damage: 0, mana: 20, effectType: 'lightning', projectile: false}
};

let moveSpeed = 3;
let moveInterval = null;
let keysPressed = {};
let currentDirection = null;
let mapEngine = null;

const NODE_TYPE = {
    GATE: 0, LOGIN: 1, LOGIC: 2, BATTLE: 3, GRIDMAP: 4, CROSS: 5, DATA: 6, GM: 7
};

const MSG = {
    LOGIN_REQ: 1001, LOGIN_RES: 1002, REGISTER_REQ: 1003, REGISTER_RES: 1004,
    LOGOUT_REQ: 1005, LOGOUT_RES: 1006, PLAYER_INFO_REQ: 3001, PLAYER_INFO_RES: 3002,
    PLAYER_MOVE_REQ: 3003, PLAYER_MOVE_RES: 3004, ATTACK_REQ: 4001, ATTACK_RES: 4002,
    SKILL_REQ: 4003, SKILL_RES: 4004, INVENTORY_REQ: 6001, INVENTORY_RES: 6002,
    ITEM_USE_REQ: 6003, ITEM_USE_RES: 6004, EQUIP_ITEM_REQ: 6005, EQUIP_ITEM_RES: 6006,
    UNEQUIP_ITEM_REQ: 6007, UNEQUIP_ITEM_RES: 6008, TASK_LIST_REQ: 7001, TASK_LIST_RES: 7002,
    TASK_ACCEPT_REQ: 7003, TASK_ACCEPT_RES: 7004, TASK_FINISH_REQ: 7005, TASK_FINISH_RES: 7006,
    MAP_LOAD_REQ: 5001, MAP_LOAD_RES: 5002, MAP_ENTITY_REQ: 5003, MAP_ENTITY_RES: 5004,
    MAP_PLAYER_ENTER: 5005, MAP_PLAYER_LEAVE: 5006, MAP_PLAYER_MOVE: 5007, MAP_PLAYER_SYNC: 5008,
    MAP_ENTITY_SYNC: 5009, PING: 9999, PONG: 9998
};

class MessageRouter {
    constructor() { this.msgToNodeType = {}; }
    register(msgId, nodeType) { this.msgToNodeType[msgId] = nodeType; }
    route(msgId) { return this.msgToNodeType[msgId] || NODE_TYPE.GATE; }
}

const msgRouter = new MessageRouter();

(function initRouter() {
    msgRouter.register(MSG.LOGIN_REQ, NODE_TYPE.LOGIN);
    msgRouter.register(MSG.LOGIN_RES, NODE_TYPE.LOGIN);
    msgRouter.register(MSG.REGISTER_REQ, NODE_TYPE.LOGIN);
    msgRouter.register(MSG.REGISTER_RES, NODE_TYPE.LOGIN);
    msgRouter.register(MSG.LOGOUT_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.LOGOUT_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.PLAYER_INFO_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.PLAYER_INFO_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.PLAYER_MOVE_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.PLAYER_MOVE_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.ATTACK_REQ, NODE_TYPE.BATTLE);
    msgRouter.register(MSG.ATTACK_RES, NODE_TYPE.BATTLE);
    msgRouter.register(MSG.SKILL_REQ, NODE_TYPE.BATTLE);
    msgRouter.register(MSG.SKILL_RES, NODE_TYPE.BATTLE);
    msgRouter.register(MSG.INVENTORY_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.INVENTORY_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.ITEM_USE_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.ITEM_USE_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.EQUIP_ITEM_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.EQUIP_ITEM_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.UNEQUIP_ITEM_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.UNEQUIP_ITEM_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_LIST_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_LIST_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_ACCEPT_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_ACCEPT_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_FINISH_REQ, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.TASK_FINISH_RES, NODE_TYPE.LOGIC);
    msgRouter.register(MSG.MAP_LOAD_REQ, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_LOAD_RES, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_ENTITY_REQ, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_ENTITY_RES, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_PLAYER_ENTER, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_PLAYER_LEAVE, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_PLAYER_MOVE, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_PLAYER_SYNC, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.MAP_ENTITY_SYNC, NODE_TYPE.GRIDMAP);
    msgRouter.register(MSG.PING, NODE_TYPE.GATE);
    msgRouter.register(MSG.PONG, NODE_TYPE.GATE);
})();

function init() {
    connectWebSocket();
    initKeyboardControls();
    initMapEngine();
    updatePlayerStats();
    setTimeout(function() {
        if (ws && ws.readyState === WebSocket.OPEN && !connected) {
            connected = true;
            document.getElementById('connStatus').className = 'status-dot online';
            document.getElementById('serverStatus').textContent = '已连接';
            document.getElementById('debugConn').textContent = '已连接';
        }
    }, 1000);
}

function initMapEngine() {
    const canvas = document.getElementById('mapCanvas');
    canvas.width = canvas.offsetWidth;
    canvas.height = canvas.offsetHeight;
    mapEngine = new MapEngine({ chunkSize: 256, viewRange: 3, tileSize: 32, renderMode: '2d' });
    mapEngine.init(canvas);
    mapEngine.setPlayerPosition({ x: playerX, y: playerY, z: 0 });
    mapEngine.loadInitialChunks();
    mapEngine.setDebugMode(true);
}

function resizeMapCanvas() {
    const canvas = document.getElementById('mapCanvas');
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width || window.innerWidth;
    canvas.height = rect.height || window.innerHeight;
    if (mapEngine && mapEngine.viewport) mapEngine.viewport.resize(canvas.width, canvas.height);
    if (mapEngine) mapEngine.forceRender();
}

function initKeyboardControls() {
    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('keyup', handleKeyUp);
}

function handleKeyDown(event) {
    if (isInBattle || !event.key) return;
    const key = event.key.toLowerCase();
    const directionMap = { 'w': 'up', 'arrowup': 'up', 's': 'down', 'arrowdown': 'down', 'a': 'left', 'arrowleft': 'left', 'd': 'right', 'arrowright': 'right' };
    const direction = directionMap[key];
    if (direction && !keysPressed[key]) {
        keysPressed[key] = true;
        currentDirection = direction;
        startContinuousMove();
    }
}

function handleKeyUp(event) {
    if (!event.key) return;
    const key = event.key.toLowerCase();
    const directionMap = { 'w': 'up', 'arrowup': 'up', 's': 'down', 'arrowdown': 'down', 'a': 'left', 'arrowleft': 'left', 'd': 'right', 'arrowright': 'right' };
    const direction = directionMap[key];
    if (direction) {
        keysPressed[key] = false;
        if (!Object.keys(keysPressed).some(k => keysPressed[k])) stopContinuousMove();
    }
}

function startContinuousMove() {
    if (moveInterval) return;
    moveInterval = setInterval(() => { if (currentDirection && !isInBattle) performMove(currentDirection); }, 50);
}

function stopContinuousMove() {
    if (moveInterval) { clearInterval(moveInterval); moveInterval = null; }
    currentDirection = null;
}

function performMove(direction) {
    const delta = direction === 'up' ? { x: 0, y: -moveSpeed } : direction === 'down' ? { x: 0, y: moveSpeed } : direction === 'left' ? { x: -moveSpeed, y: 0 } : { x: moveSpeed, y: 0 };
    const newX = playerX + delta.x;
    const newY = playerY + delta.y;
    if (newX >= 0 && newY >= 0) {
        playerX = newX; playerY = newY;
        document.getElementById('debugPos').textContent = Math.floor(playerX) + ', ' + Math.floor(playerY);
        updateCoordinateDisplay();
        if (mapEngine) mapEngine.setPlayerPosition({ x: playerX, y: playerY, z: 0 });
        sendMessage(MSG.PLAYER_MOVE_REQ, { player_id: playerID, session_id: sessionID, target_x: playerX, target_y: playerY });
        checkNPCInteraction();
    }
}

function updateCoordinateDisplay() {
    const CHUNK_SIZE = 256;
    const worldX = Math.floor(playerX); const worldY = Math.floor(playerY);
    const localX = worldX % CHUNK_SIZE; const localY = worldY % CHUNK_SIZE;
    const chunkX = Math.floor(playerX / CHUNK_SIZE); const chunkY = Math.floor(playerY / CHUNK_SIZE);
    document.getElementById('worldCoord').textContent = `世界: (${worldX}, ${worldY})`;
    document.getElementById('localCoord').textContent = `本地: (${localX}, ${localY})`;
    document.getElementById('chunkCoord').textContent = `Chunk: (${chunkX}, ${chunkY})`;
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    const addr = protocol + window.location.host + '/ws';
    if (ws) { ws.close(); ws = null; }
    ws = new WebSocket(addr);
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
        connected = true;
        document.getElementById('connStatus').className = 'status-dot online';
        document.getElementById('serverStatus').textContent = '已连接';
        document.getElementById('debugConn').textContent = '已连接';
        startHeartbeat();
    };
    ws.onmessage = (event) => handleServerMessage(event.data);
    ws.onerror = () => { connected = false; };
    ws.onclose = () => {
        connected = false;
        document.getElementById('connStatus').className = 'status-dot offline';
        document.getElementById('serverStatus').textContent = '未连接';
        document.getElementById('debugConn').textContent = '未连接';
    };
}

let heartbeatInterval = null;
function startHeartbeat() {
    if (heartbeatInterval) clearInterval(heartbeatInterval);
    heartbeatInterval = setInterval(() => { if (connected && ws) sendMessage(MSG.PING, {}); }, 30000);
}

function createPacket(msgId, nodeType, data) {
    const jsonStr = JSON.stringify(data);
    const jsonBytes = new TextEncoder().encode(jsonStr);
    const buffer = new ArrayBuffer(12 + jsonBytes.length);
    const view = new DataView(buffer);
    view.setUint32(0, jsonBytes.length, true);
    view.setUint32(4, nodeType, true);
    view.setUint32(8, msgId, true);
    for (let i = 0; i < jsonBytes.length; i++) view.setUint8(12 + i, jsonBytes[i]);
    return buffer;
}

function sendLoginMessage(msgId, data) { sendMessageToNode(msgId, NODE_TYPE.LOGIN, data); }
function sendLogicMessage(msgId, data) { sendMessageToNode(msgId, NODE_TYPE.LOGIC, data); }

function sendMessageToNode(msgId, nodeType, data) {
    const isPing = msgId === MSG.PING;
    if (!connected || !ws || ws.readyState !== WebSocket.OPEN) {
        if (!isPing) connectWebSocket();
        return;
    }
    try {
        const packet = createPacket(msgId, nodeType, data);
        ws.send(packet);
    } catch (e) { console.error('Send error:', e); }
}

function sendMessage(msgId, data) {
    const nodeType = msgRouter.route(msgId);
    sendMessageToNode(msgId, nodeType, data);
}

function handleServerMessage(data) {
    try {
        const view = new DataView(data);
        const msgId = view.getUint32(8, true);
        const msgData = new Uint8Array(data).slice(12, 12 + view.getUint32(0, true));
        const msg = JSON.parse(new TextDecoder().decode(msgData));
        switch (msgId) {
            case MSG.LOGIN_RES: handleLoginResponse(msg); break;
            case MSG.REGISTER_RES: handleRegisterResponse(msg); break;
            case MSG.LOGOUT_RES: handleLogoutResponse(msg); break;
            case MSG.PLAYER_INFO_RES: handlePlayerInfoResponse(msg); break;
            case MSG.PLAYER_MOVE_RES: handleMoveResponse(msg); break;
            case MSG.ATTACK_RES: handleAttackResponse(msg); break;
            case MSG.SKILL_RES: handleSkillResponse(msg); break;
            case MSG.INVENTORY_RES: handleInventoryResponse(msg); break;
            case MSG.ITEM_USE_RES: handleItemUseResponse(msg); break;
            case MSG.EQUIP_ITEM_RES: handleEquipResponse(msg); break;
            case MSG.UNEQUIP_ITEM_RES: handleUnequipResponse(msg); break;
            case MSG.MAP_ENTITY_RES: handleMapEntityResponse(msg); break;
            case MSG.MAP_PLAYER_SYNC: handlePlayerSync(msg); break;
            case MSG.MAP_ENTITY_SYNC: handleEntitySync(msg); break;
            case 4007: handleBattleStart(msg); break;
        }
    } catch (e) { console.error('Parse error:', e); }
}

function handleLogin() {
    const account = document.getElementById('loginAccount');
    const password = document.getElementById('loginPassword');
    if (!account || !password || !account.value || !password.value) {
        alert('请输入账号和密码');
        return;
    }
    sendLoginMessage(MSG.LOGIN_REQ, { account: account.value, password: password.value, device_id: 'web-client' });
}

function handleRegister() {
    const account = document.getElementById('regAccount').value;
    const password = document.getElementById('regPassword').value;
    const password2 = document.getElementById('regPassword2').value;
    if (!account || !password) { alert('请输入账号和密码'); return; }
    if (password !== password2) { alert('两次输入的密码不一致'); return; }
    sendLoginMessage(MSG.REGISTER_REQ, { account: account, password: password });
}

function handleLoginResponse(data) {
    if (data.result === 0) {
        isLogin = true;
        document.getElementById('loginScreen').classList.add('hidden');
        document.getElementById('gameUI').classList.remove('hidden');
        setTimeout(resizeMapCanvas, 100);
        document.getElementById('playerName').textContent = data.player_name || '测试玩家';
        document.getElementById('playerLevel').textContent = 'Lv.' + (data.level || 1);
        sessionID = data.session_id || '';
        playerID = data.player_id || 0;
        if (data.health !== undefined) { playerHealth = data.health; updatePlayerStats(); }
        addChatMessage('system', '[系统] 登录成功！欢迎来到游戏世界！');
        initTestInventory();
        setTimeout(requestInventory, 500);
    } else {
        alert('登录失败: ' + (data.message || '未知错误'));
    }
}

function initTestInventory() {
    itemConfigs = {
        1001: { id: 1001, name: '小型生命药水', type: 1, sub_type: 101, rarity: 1, level: 1, max_stack: 20, description: '恢复50点生命值', effect_type: 1, effect_value: 50 },
        1002: { id: 1002, name: '中型生命药水', type: 1, sub_type: 101, rarity: 1, level: 1, max_stack: 20, description: '恢复100点生命值', effect_type: 1, effect_value: 100 },
        2001: { id: 2001, name: '小型魔法药水', type: 1, sub_type: 101, rarity: 1, level: 1, max_stack: 20, description: '恢复30点魔法值', effect_type: 2, effect_value: 30 },
        4001: { id: 4001, name: '铁剑', type: 2, sub_type: 201, rarity: 1, level: 1, max_stack: 1, description: '普通的铁剑', attack: 10, equipment_slot: 1 },
        4003: { id: 4003, name: '皮甲', type: 2, sub_type: 202, rarity: 1, level: 1, max_stack: 1, description: '普通的皮甲', defense: 8, equipment_slot: 2 },
        5001: { id: 5001, name: '铁矿', type: 3, sub_type: 301, rarity: 1, level: 1, max_stack: 100, description: '常见的铁矿石' },
        5002: { id: 5002, name: '木材', type: 3, sub_type: 301, rarity: 1, level: 1, max_stack: 100, description: '普通的木材' },
        6001: { id: 6001, name: '神秘钥匙', type: 4, sub_type: 401, rarity: 2, level: 1, max_stack: 1, description: '开启神秘之门的钥匙' }
    };
    inventoryItems = {
        0: { itemID: 1001, count: 5, level: 1, uid: 'test_001' },
        1: { itemID: 1002, count: 2, level: 1, uid: 'test_002' },
        2: { itemID: 2001, count: 3, level: 1, uid: 'test_003' },
        3: { itemID: 4001, count: 1, level: 1, uid: 'test_004' },
        4: { itemID: 4003, count: 1, level: 1, uid: 'test_005' },
        5: { itemID: 5001, count: 10, level: 1, uid: 'test_006' },
        6: { itemID: 5002, count: 15, level: 1, uid: 'test_007' },
        7: { itemID: 6001, count: 1, level: 1, uid: 'test_008' }
    };
    updateGoldDisplay();
    renderInventory();
    renderEquipments();
}

function handleRegisterResponse(data) {
    if (data.result === 0) { alert('注册成功！请登录'); toggleLoginRegister(); }
    else { alert('注册失败: ' + (data.message || '未知错误')); }
}

function handleLogout() { sendMessage(MSG.LOGOUT_REQ, { session_id: sessionID }); }

function handleLogoutResponse(data) {
    if (data.result === 0) {
        isLogin = false;
        document.getElementById('gameUI').classList.add('hidden');
        document.getElementById('loginScreen').classList.remove('hidden');
        addChatMessage('system', '[系统] 已登出');
    }
}

function getPlayerInfo() { sendMessage(MSG.PLAYER_INFO_REQ, { player_id: playerID, session_id: sessionID }); }

function handlePlayerInfoResponse(data) {
    if (data.health !== undefined) playerHealth = data.health;
    if (data.mana !== undefined) playerMana = data.mana;
    if (data.stamina !== undefined) playerStamina = data.stamina;
    if (data.max_health !== undefined) maxHealth = data.max_health;
    if (data.max_mana !== undefined) maxMana = data.max_mana;
    if (data.max_stamina !== undefined) maxStamina = data.max_stamina;
    if (data.level !== undefined) playerLevel = data.level;
    if (data.exp !== undefined) playerExp = data.exp;
    if (data.strength !== undefined) playerStrength = data.strength;
    if (data.agility !== undefined) playerAgility = data.agility;
    if (data.intelligence !== undefined) playerIntelligence = data.intelligence;
    if (data.defense !== undefined) playerDefense = data.defense;
    updatePlayerStats();
}

function updatePlayerStats() {
    playerHealth = Math.max(0, Math.min(maxHealth, playerHealth));
    playerMana = Math.max(0, Math.min(maxMana, playerMana));
    playerStamina = Math.max(0, Math.min(maxStamina, playerStamina));
    
    document.getElementById('healthVal').textContent = Math.round(playerHealth) + '/' + maxHealth;
    document.getElementById('manaVal').textContent = Math.round(playerMana) + '/' + maxMana;
    document.getElementById('staminaVal').textContent = Math.round(playerStamina) + '/' + maxStamina;
    
    const healthPercent = (playerHealth / maxHealth) * 100;
    const manaPercent = (playerMana / maxMana) * 100;
    const staminaPercent = (playerStamina / maxStamina) * 100;
    
    document.getElementById('healthBar').style.width = healthPercent + '%';
    document.getElementById('manaBar').style.width = manaPercent + '%';
    document.getElementById('staminaBar').style.width = staminaPercent + '%';
    
    document.getElementById('playerHealth').textContent = Math.round(playerHealth) + '/' + maxHealth;
    document.getElementById('playerHealthBar').style.width = healthPercent + '%';
    document.getElementById('playerHealthBar').className = 'battle-health-fill' + (playerHealth < maxHealth * 0.3 ? ' danger' : '');
    
    document.getElementById('statStrength').textContent = playerStrength;
    document.getElementById('statAgility').textContent = playerAgility;
    document.getElementById('statIntelligence').textContent = playerIntelligence;
    document.getElementById('statDefense').textContent = playerDefense;
    
    document.getElementById('statusLevel').textContent = 'Lv.' + playerLevel;
    document.getElementById('statusExp').textContent = playerExp;
    document.getElementById('statusGold').textContent = playerGold;
}

function toggleLoginRegister() {
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const switchText = document.getElementById('loginSwitch');
    if (loginForm.classList.contains('hidden')) {
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
        switchText.textContent = '还没有账号？点击注册';
    } else {
        loginForm.classList.add('hidden');
        registerForm.classList.remove('hidden');
        switchText.textContent = '已有账号？点击登录';
    }
}

function movePlayer(direction) { if (!isInBattle) performMove(direction); }

function checkNPCInteraction() {
    const npcX = 180, npcY = 120;
    const dist = Math.sqrt(Math.pow(playerX - npcX, 2) + Math.pow(playerY - npcY, 2));
    if (dist < 30) showNPCDialog();
}

function showNPCDialog() {
    document.getElementById('npcDialog').classList.remove('hidden');
    document.getElementById('dialogNpcName').textContent = '🧙 老法师';
    document.getElementById('dialogText').textContent = '欢迎来到翡翠森林，勇敢的冒险者。我有一个任务交给你...';
    showDialogOptions([{ text: '什么任务？', next: 1 }, { text: '我很忙，下次再说', next: -1 }]);
}

function showDialogOptions(options) {
    const container = document.getElementById('dialogOptions');
    container.innerHTML = '';
    options.forEach(opt => {
        const div = document.createElement('div');
        div.className = 'dialog-option';
        div.textContent = opt.text;
        div.onclick = () => { if (opt.next === -1) document.getElementById('npcDialog').classList.add('hidden'); else continueDialog(opt.next); };
        container.appendChild(div);
    });
}

function continueDialog(step) {
    const dialogs = [
        { text: '我需要你去森林深处清除一些哥布林，它们已经骚扰村民太久了。', options: [{ text: '我接受这个任务！', next: 2 }, { text: '我需要准备一下', next: -1 }] },
        { text: '太好了！完成任务后回来找我，我会给你丰厚的奖励。', options: [{ text: '好的，我这就去！', next: -1 }] }
    ];
    const dialog = dialogs[step - 1];
    if (dialog) { document.getElementById('dialogText').textContent = dialog.text; showDialogOptions(dialog.options); }
}

function useSkill(skillId) {
    const skill = getSkillConfig(skillId);
    if (!skill) return;
    if (isSkillOnCooldown(skillId)) {
        addBattleLog('system', `技能冷却中，剩余 ${getSkillCooldownRemaining(skillId)} 秒`);
        return;
    }
    if (playerMana < skill.mana) { addBattleLog('system', '魔法值不足！'); return; }
    playerMana -= skill.mana;
    setSkillCooldown(skillId);
    playSkillEffect(skillId, skill);
    
    if (skillId >= 100) {
        if (skillId === 100) {
            playerHealth = Math.min(MAX_HEALTH, playerHealth - skill.damage);
            addBattleLog('heal', `使用 ${skill.name}，恢复了 ${-skill.damage} 点生命值`);
            showDamageNumber(300, 200, -skill.damage, 'heal');
        } else addBattleLog('skill', `使用 ${skill.name}`);
    } else if (isInBattle && currentTarget) {
        monsterHealth = Math.max(0, monsterHealth - skill.damage);
        addBattleLog('damage', `使用 ${skill.name} 对 ${currentTarget} 造成了 ${skill.damage} 点伤害`);
        showDamageNumber(700, 300, skill.damage, 'damage');
        updateMonsterHealth();
        if (monsterHealth <= 0) endBattle(true);
    }
    updatePlayerStats();
    if (connected && playerID > 0) {
        sendMessage(MSG.SKILL_REQ, { battle_id: battleID, player_id: playerID, skill_id: skillId, target_id: currentTarget || 0 });
    }
}

function playSkillEffect(skillId, skill) {
    const battleUI = document.getElementById('battleUI');
    const isBattleVisible = battleUI && !battleUI.classList.contains('hidden');
    
    let playerPos, monsterPos;
    if (isBattleVisible) {
        playerPos = { x: 60, y: 30 };
        monsterPos = { x: 700, y: 30 };
    } else {
        const playerMarker = document.getElementById('playerMarker');
        if (playerMarker && playerMarker.style.display !== 'none') {
            const rect = playerMarker.getBoundingClientRect();
            playerPos = { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
        } else {
            playerPos = { x: window.innerWidth / 2 - 30, y: window.innerHeight / 2 - 30 };
        }
        monsterPos = { x: playerPos.x + 100, y: playerPos.y };
    }
    
    createSkillEffect(playerPos.x, playerPos.y, skill.effectType);
    
    if (skill.projectile && skillId < 100) {
        if (isSkillOnCooldown(skillId)) return;
        createProjectile(playerPos.x, playerPos.y, monsterPos.x, monsterPos.y, skill.effectType);
        setTimeout(() => createSkillEffect(monsterPos.x, monsterPos.y, skill.effectType), 400);
    }
    
    if (skillId === 100) {
        createSkillEffect(playerPos.x, playerPos.y, skill.effectType);
    }
}

function createSkillEffect(x, y, effectType) {
    const effect = document.createElement('div');
    effect.className = `skill-effect skill-effect-${effectType}`;
    effect.style.left = x + 'px';
    effect.style.top = y + 'px';
    effect.style.width = '60px';
    effect.style.height = '60px';
    document.body.appendChild(effect);
    setTimeout(() => effect.remove(), 600);
}

function createProjectile(startX, startY, endX, endY, effectType) {
    const projectile = document.createElement('div');
    projectile.className = `skill-projectile projectile-${effectType}`;
    projectile.style.left = startX + 'px';
    projectile.style.top = startY + 'px';
    document.body.appendChild(projectile);
    const duration = 400;
    const startTime = Date.now();
    function animate() {
        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / duration, 1);
        projectile.style.left = startX + (endX - startX) * progress + 'px';
        projectile.style.top = startY + (endY - startY) * progress + 'px';
        if (progress < 1) requestAnimationFrame(animate);
        else projectile.remove();
    }
    requestAnimationFrame(animate);
}

function showDamageNumber(x, y, value, type) {
    const damageNum = document.createElement('div');
    damageNum.className = `damage-number ${type}`;
    damageNum.textContent = (type === 'damage' ? '-' : '+') + value;
    damageNum.style.left = x + 'px';
    damageNum.style.top = y + 'px';
    document.body.appendChild(damageNum);
    setTimeout(() => damageNum.remove(), 1000);
}

function handleSkillResponse(data) {
    if (data.result === 0) {
        addChatMessage('system', `技能释放成功！`);
        if (data.damage) addBattleLog('damage', `技能造成了 ${data.damage} 点伤害`);
        if (data.heal) { playerHealth = Math.min(MAX_HEALTH, playerHealth + data.heal); addBattleLog('heal', `恢复了 ${data.heal} 点生命值`); }
        if (data.mana_cost !== undefined) playerMana = Math.max(0, playerMana - data.mana_cost);
        if (data.health !== undefined) playerHealth = data.health;
        if (data.mana !== undefined) playerMana = data.mana;
        updatePlayerStats();
    } else addChatMessage('system', `技能释放失败: ${data.message || '未知错误'}`);
}

function startBattle(monsterId) {
    const monsters = { 1: { name: '哥布林', health: 80, level: 1 }, 2: { name: '幼龙', health: 200, level: 5 } };
    const monster = monsters[monsterId];
    if (!monster) return;
    isInBattle = true;
    currentTarget = monster.name;
    monsterHealth = monster.health;
    battleID = Date.now();
    document.getElementById('debugBattle').textContent = '是';
    document.getElementById('battleUI').classList.remove('hidden');
    document.getElementById('monsterName').textContent = monster.name;
    document.getElementById('monsterHealth').textContent = monster.health + '/' + monster.health;
    document.getElementById('monsterHealthBar').style.width = '100%';
    addBattleLog('system', `进入战斗！目标: ${monster.name} Lv.${monster.level}`);
}

function updateMonsterHealth() {
    document.getElementById('monsterHealth').textContent = monsterHealth + '/80';
    document.getElementById('monsterHealthBar').style.width = (monsterHealth / 80 * 100) + '%';
    document.getElementById('monsterHealthBar').className = 'battle-health-fill' + (monsterHealth < 24 ? ' danger' : '');
}

function endBattle(victory) {
    isInBattle = false;
    document.getElementById('battleUI').classList.add('hidden');
    document.getElementById('debugBattle').textContent = '否';
    if (victory) {
        addBattleLog('system', `胜利！击败了 ${currentTarget}！`);
        playerHealth = Math.min(MAX_HEALTH, playerHealth + 10);
        updatePlayerStats();
    } else addBattleLog('system', '战斗失败...');
    currentTarget = null;
    battleID = 0;
}

function openAttackMode() { if (!isInBattle) startBattle(1); }

function sendChat() {
    const msg = document.getElementById('chatInput').value.trim();
    if (!msg) return;
    sendMessage(MSG.CHAT_REQ, { message: msg });
    addChatMessage('player', `[你] ${msg}`);
    document.getElementById('chatInput').value = '';
}

function addChatMessage(type, msg) {
    const container = document.getElementById('chatMessages');
    const div = document.createElement('div');
    div.className = 'chat-message ' + type;
    div.textContent = msg;
    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
}

function handleBattleStart(data) {
    if (data.message === 'battle_start') { alert('你进入了战斗！'); startBattle(1); }
}

function handleMapEntityResponse(data) { if (data.entities) updateNearbyMonsters(data.entities); }
function handlePlayerSync(data) {}
function handleEntitySync(data) { if (data.entities) updateNearbyMonsters(data.entities); }

function updateNearbyMonsters(entities) {
    const panel = document.querySelector('.monsters-panel');
    if (!panel) return;
    let container = panel.querySelector('.monsters-container');
    if (!container) { container = document.createElement('div'); container.className = 'monsters-container'; panel.appendChild(container); }
    const monsterItems = entities.filter(e => e.entity_type === 1 || e.EntityType === 1);
    container.innerHTML = '';
    monsterItems.forEach((monster, index) => {
        const item = document.createElement('div');
        item.className = 'monster-item';
        item.innerHTML = `<span class="monster-name">${monster.name || monster.Name}</span><span class="monster-level">Lv.${monster.level || 1}</span>`;
        item.onclick = () => startBattle(index + 1);
        container.appendChild(item);
    });
}

function addBattleLog(type, msg) {
    const container = document.getElementById('battleLog');
    const div = document.createElement('div');
    div.className = 'battle-log-entry';
    if (type === 'damage') div.innerHTML = `<span class="battle-log-damage">${msg}</span>`;
    else if (type === 'heal') div.innerHTML = `<span class="battle-log-heal">${msg}</span>`;
    else if (type === 'skill') div.innerHTML = `<span class="battle-log-skill">${msg}</span>`;
    else div.textContent = msg;
    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
}

function handleMoveResponse(data) {
    if (data.x !== undefined || data.pos_x !== undefined) {
        playerX = data.x !== undefined ? data.x : data.pos_x;
        playerY = data.y !== undefined ? data.y : (data.pos_y !== undefined ? data.pos_y : playerY);
        document.getElementById('debugPos').textContent = Math.floor(playerX) + ', ' + Math.floor(playerY);
        updateCoordinateDisplay();
        if (mapEngine) mapEngine.setPlayerPosition({ x: playerX, y: playerY, z: 0 });
    }
}

function handleAttackResponse(data) {
    if (data.damage) {
        playerHealth = Math.max(0, playerHealth - data.damage);
        addBattleLog('damage', `${currentTarget} 对你造成了 ${data.damage} 点伤害`);
        updatePlayerStats();
        if (playerHealth <= 0) endBattle(false);
    }
}

function requestInventory() { sendMessage(MSG.INVENTORY_REQ, { player_id: playerID, session_id: sessionID }); }

function handleInventoryResponse(data) {
    if (data.result === 0) {
        if (data.gold !== undefined) { playerGold = data.gold; updateGoldDisplay(); }
        if (data.items) {
            inventoryItems = {};
            data.items.forEach(item => { inventoryItems[item.slot] = { itemID: item.item_id, count: item.count, level: item.level || 0, uid: item.uid }; });
            renderInventory();
        }
        if (data.equipments) {
            equipments = {};
            data.equipments.forEach(equip => { equipments[equip.slot] = { itemID: equip.item_id, level: equip.level || 0 }; });
            renderEquipments();
        }
        if (data.capacity !== undefined) inventoryCapacity = data.capacity;
        if (data.item_configs) data.item_configs.forEach(config => { itemConfigs[config.id] = config; });
    } else console.error('获取背包失败:', data.message);
}

function useItem(slot) {
    const item = inventoryItems[slot];
    if (!item) return;
    const config = itemConfigs[item.itemID];
    if (!config) { console.error('物品配置不存在:', item.itemID); return; }
    if (config.type !== 1) { addChatMessage('system', '[系统] 该物品不能使用'); return; }
    sendMessage(MSG.ITEM_USE_REQ, { player_id: playerID, session_id: sessionID, item_id: item.itemID, position: slot });
}

function handleItemUseResponse(data) {
    if (data.result === 0) {
        if (data.slot !== undefined) {
            if (data.count > 0) inventoryItems[data.slot].count = data.count;
            else delete inventoryItems[data.slot];
            renderInventory();
        }
        if (data.effect) applyItemEffect(data.effect);
        addChatMessage('system', '[系统] 使用物品成功');
    } else addChatMessage('system', '[系统] 使用物品失败: ' + (data.message || '未知错误'));
}

function applyItemEffect(effect) {
    switch (effect.type) {
        case 1: playerHealth = Math.min(MAX_HEALTH, playerHealth + effect.value); addChatMessage('system', `[系统] 恢复了 ${effect.value} 点生命值`); break;
        case 2: playerMana = Math.min(MAX_MANA, playerMana + effect.value); addChatMessage('system', `[系统] 恢复了 ${effect.value} 点魔法值`); break;
        case 3: addChatMessage('system', `[系统] 力量+${effect.value}`); break;
        case 4: addChatMessage('system', `[系统] 敏捷+${effect.value}`); break;
        case 5: addChatMessage('system', `[系统] 智力+${effect.value}`); break;
    }
    updatePlayerStats();
}

function equipItem(slot) {
    const item = inventoryItems[slot];
    if (!item) return;
    const config = itemConfigs[item.itemID];
    if (!config) { console.error('物品配置不存在:', item.itemID); return; }
    if (config.category !== 2) { addChatMessage('system', '[系统] 该物品不是装备'); return; }
    sendMessage(MSG.EQUIP_ITEM_REQ, { player_id: playerID, session_id: sessionID, slot: slot });
}

function handleEquipResponse(data) {
    if (data.result === 0) {
        if (data.old_slot !== undefined) delete inventoryItems[data.old_slot];
        if (data.new_item) inventoryItems[data.new_slot] = { itemID: data.new_item.item_id, count: data.new_item.count, level: data.new_item.level || 0, uid: data.new_item.uid };
        if (data.equipment) equipments[data.equipment.slot] = { itemID: data.equipment.item_id, level: data.equipment.level || 0 };
        renderInventory();
        renderEquipments();
        addChatMessage('system', '[系统] 装备成功');
    } else addChatMessage('system', '[系统] 装备失败: ' + (data.message || '未知错误'));
}

function unequipItem(slot) {
    if (!equipments[slot]) return;
    sendMessage(MSG.UNEQUIP_ITEM_REQ, { player_id: playerID, session_id: sessionID, slot: slot });
}

function handleUnequipResponse(data) {
    if (data.result === 0) {
        delete equipments[data.slot];
        if (data.item) inventoryItems[data.item.slot] = { itemID: data.item.item_id, count: data.item.count, level: data.item.level || 0, uid: data.item.uid };
        renderInventory();
        renderEquipments();
        addChatMessage('system', '[系统] 卸下装备成功');
    } else addChatMessage('system', '[系统] 卸下装备失败: ' + (data.message || '未知错误'));
}

function updateGoldDisplay() {
    const goldDisplay = document.querySelector('.gold-value');
    if (goldDisplay) goldDisplay.textContent = '💰 ' + playerGold.toLocaleString();
}

function renderInventory() {
    const grid = document.getElementById('inventoryGrid');
    if (!grid) return;
    grid.innerHTML = '';
    for (let i = 0; i < 16; i++) {
        const item = inventoryItems[i];
        const cell = document.createElement('div');
        cell.className = 'inventory-item';
        if (item) {
            const config = itemConfigs[item.itemID];
            const icon = ITEM_ICONS[item.itemID] || '📦';
            cell.innerHTML = `<div class="item-icon">${icon}</div>${item.count > 1 ? `<div class="item-count">${item.count}</div>` : ''}`;
            const rarityColors = { 1: '#9ca3af', 2: '#22c55e', 3: '#3b82f6', 4: '#a855f7', 5: '#f59e0b' };
            if (config && rarityColors[config.rarity]) cell.style.borderColor = rarityColors[config.rarity];
            cell.onclick = () => useItem(i);
        } else cell.innerHTML = '<div class="item-icon" style="opacity: 0.3;">📦</div>';
        grid.appendChild(cell);
    }
}

function renderEquipments() {
    const equipPanel = document.getElementById('equipmentPanel');
    if (!equipPanel) return;
    equipPanel.innerHTML = '';
    for (let i = 1; i <= 7; i++) {
        const equip = equipments[i];
        const slotName = EQUIP_SLOT_NAMES[i];
        const div = document.createElement('div');
        div.className = 'equip-slot';
        if (equip) {
            const config = itemConfigs[equip.itemID];
            const icon = ITEM_ICONS[equip.itemID] || '📦';
            div.innerHTML = `<span class="equip-slot-name">${slotName}</span><span class="equip-icon">${icon}</span>`;
            div.onclick = () => showEquipTooltip(equip, config, i);
        } else div.innerHTML = `<span class="equip-slot-name">${slotName}</span><span class="equip-icon" style="opacity: 0.3;">📦</span>`;
        equipPanel.appendChild(div);
    }
}

function showItemTooltip(item, config, slot) {
    if (!config) return;
    const tooltip = document.createElement('div');
    tooltip.className = 'item-tooltip';
    const rarityColors = { 1: '#9ca3af', 2: '#22c55e', 3: '#3b82f6', 4: '#a855f7', 5: '#f59e0b' };
    const color = rarityColors[config.rarity] || '#fff';
    tooltip.innerHTML = `
        <div style="color: ${color}; font-weight: bold;">${config.name}</div>
        <div style="font-size: 12px; color: #888;">Lv.${config.level || 1}</div>
        <div style="margin-top: 5px; font-size: 11px; color: #aaa;">${config.description}</div>
        ${config.attack ? `<div style="font-size: 11px; color: #ef4444;">攻击 +${config.attack}</div>` : ''}
        ${config.defense ? `<div style="font-size: 11px; color: #3b82f6;">防御 +${config.defense}</div>` : ''}
        ${config.strength ? `<div style="font-size: 11px; color: #f59e0b;">力量 +${config.strength}</div>` : ''}
        <div style="margin-top: 10px; display: flex; gap: 10px;">
            ${config.category === 1 ? `<button class="tooltip-btn" onclick="useItem(${slot}); closeTooltip();">使用</button>` : ''}
            ${config.category === 2 ? `<button class="tooltip-btn" onclick="equipItem(${slot}); closeTooltip();">装备</button>` : ''}
            <button class="tooltip-btn" onclick="closeTooltip()">关闭</button>
        </div>
    `;
    tooltip.style.cssText = 'position: fixed; left: 50%; top: 50%; transform: translate(-50%, -50%); background: rgba(0,0,0,0.95); padding: 15px; border-radius: 10px; border: 1px solid rgba(255,255,255,0.2); z-index: 1000; min-width: 200px;';
    document.body.appendChild(tooltip);
}

function showEquipTooltip(equip, config, slot) {
    if (!config) return;
    const tooltip = document.createElement('div');
    tooltip.className = 'item-tooltip';
    const rarityColors = { 1: '#9ca3af', 2: '#22c55e', 3: '#3b82f6', 4: '#a855f7', 5: '#f59e0b' };
    const color = rarityColors[config.rarity] || '#fff';
    tooltip.innerHTML = `
        <div style="color: ${color}; font-weight: bold;">${config.name}</div>
        <div style="font-size: 12px; color: #888;">Lv.${config.level || 1}</div>
        <div style="margin-top: 5px; font-size: 11px; color: #aaa;">${config.description}</div>
        ${config.attack ? `<div style="font-size: 11px; color: #ef4444;">攻击 +${config.attack}</div>` : ''}
        ${config.defense ? `<div style="font-size: 11px; color: #3b82f6;">防御 +${config.defense}</div>` : ''}
        ${config.strength ? `<div style="font-size: 11px; color: #f59e0b;">力量 +${config.strength}</div>` : ''}
        ${config.agility ? `<div style="font-size: 11px; color: #22c55e;">敏捷 +${config.agility}</div>` : ''}
        ${config.intelligence ? `<div style="font-size: 11px; color: #a855f7;">智力 +${config.intelligence}</div>` : ''}
        <div style="margin-top: 10px;">
            <button class="tooltip-btn" onclick="unequipItem(${slot}); closeTooltip();">卸下</button>
            <button class="tooltip-btn" onclick="closeTooltip()">关闭</button>
        </div>
    `;
    tooltip.style.cssText = 'position: fixed; left: 50%; top: 50%; transform: translate(-50%, -50%); background: rgba(0,0,0,0.95); padding: 15px; border-radius: 10px; border: 1px solid rgba(255,255,255,0.2); z-index: 1000; min-width: 200px;';
    document.body.appendChild(tooltip);
}

function closeTooltip() {
    const tooltip = document.querySelector('.item-tooltip');
    if (tooltip) tooltip.remove();
}

function clearCache() { localStorage.clear(); alert('缓存已清除'); }

window.onload = init;

window.handleLogin = handleLogin;
window.handleRegister = handleRegister;
window.toggleLoginRegister = toggleLoginRegister;
window.handleLogout = handleLogout;
window.movePlayer = movePlayer;
window.useSkill = useSkill;
window.startBattle = startBattle;
window.openAttackMode = openAttackMode;
window.sendChat = sendChat;
window.clearCache = clearCache;
window.useItem = useItem;
window.equipItem = equipItem;
window.unequipItem = unequipItem;
window.closeTooltip = closeTooltip;