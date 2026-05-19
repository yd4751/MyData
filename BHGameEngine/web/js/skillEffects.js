// 技能特效系统

// 播放技能特效
function playSkillEffect(skillId, skill) {
    const battleUI = document.getElementById('battleUI');
    if (!battleUI || battleUI.classList.contains('hidden')) return;
    
    const playerPos = { x: 60, y: 30 };
    const monsterPos = { x: 700, y: 30 };
    
    // 播放施法者特效
    createSkillEffect(playerPos.x, playerPos.y, skill.effectType, 'cast');
    
    // 如果有弹道，创建弹道效果
    if (skill.projectile !== undefined && skill.projectile !== null && skillId < 100 && isInBattle) {
        createProjectile(playerPos.x, playerPos.y, monsterPos.x, monsterPos.y, skill.effectType);
        
        setTimeout(() => {
            createSkillEffect(monsterPos.x, monsterPos.y, skill.effectType, 'hit');
        }, 400);
    } else if (skillId < 100 && isInBattle) {
        setTimeout(() => {
            createSkillEffect(monsterPos.x, monsterPos.y, skill.effectType, 'hit');
        }, 200);
    }
    
    if (skillId >= 100) {
        setTimeout(() => {
            createSkillEffect(playerPos.x, playerPos.y, skill.effectType, 'hit');
        }, 100);
    }
}

// 创建技能特效
function createSkillEffect(x, y, effectType, effectMode = 'hit') {
    const effect = document.createElement('div');
    const effectClass = effectMode === 'cast' ? 'skill-effect-cast' : 'skill-effect-hit';
    effect.className = `skill-effect skill-effect-${effectType} ${effectClass}`;
    effect.style.left = x + 'px';
    effect.style.top = y + 'px';
    effect.style.width = effectMode === 'cast' ? '50px' : '70px';
    effect.style.height = effectMode === 'cast' ? '50px' : '70px';
    
    createParticles(x, y, effectType);
    
    document.body.appendChild(effect);
    
    const duration = effectMode === 'cast' ? 300 : 500;
    setTimeout(() => {
        effect.remove();
    }, duration);
}

// 创建粒子效果
function createParticles(x, y, effectType) {
    const particleCount = 8;
    const colors = {
        fire: ['#ff6b35', '#f7931e', '#ff4500'],
        ice: ['#00d4ff', '#3b82f6', '#1e40af'],
        lightning: ['#fbbf24', '#facc15', '#eab308'],
        heal: ['#22c55e', '#16a34a', '#15803d'],
        shield: ['#8b5cf6', '#6366f1', '#4f46e5'],
        stun: ['#f59e0b', '#d97706', '#b45309'],
        silence: ['#6b7280', '#4b5563', '#374151'],
        knockback: ['#06b6d4', '#0891b2', '#0e7490'],
        normal: ['#9ca3af', '#6b7280', '#4b5563']
    };
    
    const effectColors = colors[effectType] || colors.normal;
    
    for (let i = 0; i < particleCount; i++) {
        const particle = document.createElement('div');
        particle.className = 'skill-particle';
        particle.style.left = x + 'px';
        particle.style.top = y + 'px';
        particle.style.width = '6px';
        particle.style.height = '6px';
        particle.style.background = effectColors[Math.floor(Math.random() * effectColors.length)];
        particle.style.borderRadius = '50%';
        
        const angle = (Math.PI * 2 * i) / particleCount;
        const distance = 30 + Math.random() * 20;
        const endX = x + Math.cos(angle) * distance;
        const endY = y + Math.sin(angle) * distance;
        
        particle.style.setProperty('--endX', endX + 'px');
        particle.style.setProperty('--endY', endY + 'px');
        
        document.body.appendChild(particle);
        
        setTimeout(() => {
            particle.remove();
        }, 400);
    }
}

// 创建技能弹道
function createProjectile(startX, startY, endX, endY, effectType) {
    const projectile = document.createElement('div');
    projectile.className = `skill-projectile projectile-${effectType}`;
    projectile.style.left = startX + 'px';
    projectile.style.top = startY + 'px';
    
    projectile.style.setProperty('--startX', startX + 'px');
    projectile.style.setProperty('--startY', startY + 'px');
    projectile.style.setProperty('--endX', endX + 'px');
    projectile.style.setProperty('--endY', endY + 'px');
    
    document.body.appendChild(projectile);
    
    const duration = 400;
    const startTime = Date.now();
    
    function animate() {
        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / duration, 1);
        
        const currentX = startX + (endX - startX) * progress;
        const currentY = startY + (endY - startY) * progress;
        
        projectile.style.left = currentX + 'px';
        projectile.style.top = currentY + 'px';
        
        if (progress < 1) {
            requestAnimationFrame(animate);
        } else {
            projectile.remove();
        }
    }
    
    requestAnimationFrame(animate);
}

// 显示伤害/治疗数字
function showDamageNumber(x, y, value, type) {
    const damageNum = document.createElement('div');
    damageNum.className = `damage-number ${type}`;
    damageNum.textContent = (type === 'damage' ? '-' : '+') + value;
    damageNum.style.left = x + 'px';
    damageNum.style.top = y + 'px';
    
    document.body.appendChild(damageNum);
    
    setTimeout(() => {
        damageNum.remove();
    }, 1000);
}