// 技能冷却系统

// 技能配置扩展（添加冷却时间）
const SKILL_CONFIG_EXTENDED = {
    1: {name: '普通攻击', damage: 15, mana: 0, effectType: 'normal', projectile: false, cooldown: 1},
    2: {name: '火球术', damage: 30, mana: 10, effectType: 'fire', projectile: true, cooldown: 3},
    3: {name: '冰霜箭', damage: 25, mana: 8, effectType: 'ice', projectile: true, cooldown: 2},
    4: {name: '闪电链', damage: 20, mana: 12, effectType: 'lightning', projectile: true, cooldown: 4},
    100: {name: '治疗术', damage: -30, mana: 15, effectType: 'heal', projectile: false, cooldown: 5},
    101: {name: '护盾', damage: 0, mana: 10, effectType: 'shield', projectile: false, cooldown: 6},
    102: {name: '加速', damage: 0, mana: 5, effectType: 'lightning', projectile: false, cooldown: 8},
    103: {name: '瞬移', damage: 0, mana: 20, effectType: 'lightning', projectile: false, cooldown: 10}
};

// 技能冷却状态
let skillCooldowns = {};

// 检查技能是否在冷却中
function isSkillOnCooldown(skillId) {
    if (!skillCooldowns[skillId]) return false;
    return skillCooldowns[skillId] > Date.now();
}

// 获取技能剩余冷却时间（秒）
function getSkillCooldownRemaining(skillId) {
    if (!skillCooldowns[skillId]) return 0;
    const remaining = Math.ceil((skillCooldowns[skillId] - Date.now()) / 1000);
    return Math.max(0, remaining);
}

// 设置技能冷却
function setSkillCooldown(skillId) {
    const skill = SKILL_CONFIG_EXTENDED[skillId];
    if (!skill || !skill.cooldown) return;
    
    skillCooldowns[skillId] = Date.now() + (skill.cooldown * 1000);
    updateSkillCooldownDisplay(skillId);
    
    const cooldownInterval = setInterval(() => {
        const remaining = getSkillCooldownRemaining(skillId);
        updateSkillCooldownDisplay(skillId, remaining);
        
        if (remaining <= 0) {
            clearInterval(cooldownInterval);
            removeSkillCooldownDisplay(skillId);
        }
    }, 1000);
}

// 更新技能冷却显示
function updateSkillCooldownDisplay(skillId, remaining = null) {
    if (remaining === null) {
        remaining = getSkillCooldownRemaining(skillId);
    }
    
    const skillElement = document.querySelector(`[data-skill-id="${skillId}"]`);
    if (!skillElement) return;
    
    let overlay = skillElement.querySelector('.skill-cooldown-overlay');
    
    if (remaining > 0) {
        if (!overlay) {
            overlay = document.createElement('div');
            overlay.className = 'skill-cooldown-overlay';
            skillElement.appendChild(overlay);
        }
        
        overlay.innerHTML = `
            <div class="skill-cooldown-time">${remaining}</div>
        `;
        skillElement.classList.add('on-cooldown');
    } else {
        if (overlay) {
            overlay.remove();
        }
        skillElement.classList.remove('on-cooldown');
    }
}

// 移除技能冷却显示
function removeSkillCooldownDisplay(skillId) {
    const skillElement = document.querySelector(`[data-skill-id="${skillId}"]`);
    if (!skillElement) return;
    
    const overlay = skillElement.querySelector('.skill-cooldown-overlay');
    if (overlay) {
        overlay.remove();
    }
    skillElement.classList.remove('on-cooldown');
}

// 获取技能配置（优先使用扩展配置）
function getSkillConfig(skillId) {
    return SKILL_CONFIG_EXTENDED[skillId] || SKILL_CONFIG[skillId];
}