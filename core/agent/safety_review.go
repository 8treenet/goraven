package agent

import (
	"fmt"
	"regexp"
	"strings"

	"goraven/config"
)

func getSafetyPrefix() string {
	if config.Get().GetLanguage() == "zh" {
		return "安全审核拒绝: "
	}
	return "Safety review rejected: "
}

type safetyRule struct {
	pattern *regexp.Regexp
	descZh  string
	descEn  string
}

// baseSafetyRules 基础安全审核规则，适用于所有 Agent。
// 禁止最危险的操作：系统摧毁、数据损坏、权限绕过等。
var baseSafetyRules = []safetyRule{
	// 1. 递归强制删除根目录，防止系统不可逆损坏
	{
		pattern: regexp.MustCompile(`\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+/(\s|$|;|\||&)`),
		descZh:  "禁止递归强制删除根目录: rm -rf /",
		descEn:  "Forbidden recursive force delete of root directory: rm -rf /",
	},

	// 2. 递归强制删除根下所有内容
	{
		pattern: regexp.MustCompile(`\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+/\*`),
		descZh:  "禁止递归强制删除根目录下所有内容: rm -rf /*",
		descEn:  "Forbidden recursive force delete all content under root: rm -rf /*",
	},

	// 3. fork 炸弹，耗尽系统进程资源
	{
		pattern: regexp.MustCompile(`:\(\)\s*\{\s*:\|\:&\s*\}\s*;`),
		descZh:  "禁止 fork 炸弹: :(){ :|:& };:",
		descEn:  "Forbidden fork bomb: :(){ :|:& };:",
	},

	// 4. 访问系统密码文件
	{
		pattern: regexp.MustCompile(`/etc/(?:shadow|passwd)\b`),
		descZh:  "禁止访问系统密码文件: /etc/shadow, /etc/passwd",
		descEn:  "Forbidden access to system password files: /etc/shadow, /etc/passwd",
	},

	// 5. 关机/重启
	{
		pattern: regexp.MustCompile(`\b(?:reboot|shutdown|halt|poweroff|init\s+[06])\b`),
		descZh:  "禁止关机/重启操作: reboot/shutdown/halt/poweroff",
		descEn:  "Forbidden shutdown/reboot: reboot/shutdown/halt/poweroff",
	},

	// 6. 直接访问裸块设备
	{
		pattern: regexp.MustCompile(`/dev/sd[a-z]|/dev/nvme|/dev/xvd|/dev/vd[a-z]|/dev/hd[a-z]`),
		descZh:  "禁止直接访问裸块设备文件",
		descEn:  "Forbidden direct access to raw block device files",
	},

	// 7. dd 磁盘转储命令
	{
		pattern: regexp.MustCompile(`\bdd\s+if=`),
		descZh:  "禁止 dd 磁盘转储命令",
		descEn:  "Forbidden dd disk dump command",
	},

	// 8. mkfs 格式化文件系统
	{
		pattern: regexp.MustCompile(`\bmkfs\.?\w*\b`),
		descZh:  "禁止 mkfs 创建文件系统（格式化）",
		descEn:  "Forbidden mkfs filesystem creation (formatting)",
	},

	// 9. netcat 反弹 shell
	{
		pattern: regexp.MustCompile(`\bnc?\s+.*-e\s+/bin/(?:ba|)sh\b|\bnetcat\s+.*-e\s+/bin/(?:ba|)sh\b`),
		descZh:  "禁止 netcat 反弹 shell: nc -e /bin/sh",
		descEn:  "Forbidden netcat reverse shell: nc -e /bin/sh",
	},

	// 10. iptables 清空防火墙规则
	{
		pattern: regexp.MustCompile(`\biptables\s+-F\b|\biptables\s+-P\s+\w+\s+ACCEPT`),
		descZh:  "禁止清空或开放所有 iptables 防火墙规则",
		descEn:  "Forbidden flushing or opening all iptables firewall rules",
	},
}

// extraStrictRules 额外严格规则，仅适用于 MainAgent。
// 禁止 sudo、curl|sh、敏感文件访问、包安装等，
// 在 baseSafetyRules 基础上进一步加强安全管控。
var extraStrictRules = []safetyRule{
	// 11. sudo 提权
	{
		pattern: regexp.MustCompile(`\bsudo\b`),
		descZh:  "禁止使用 sudo 提权执行命令",
		descEn:  "Forbidden sudo privilege escalation",
	},

	// 12. chmod 777 设置世界可写权限
	{
		pattern: regexp.MustCompile(`\bchmod\s+(-R\s+)?777\b`),
		descZh:  "禁止 chmod 777 设置全开放权限",
		descEn:  "Forbidden chmod 777 world-writable permission",
	},

	// 13. curl 管道到 shell
	{
		pattern: regexp.MustCompile(`\bcurl\s+.*\|\s*(?:ba)?sh\b`),
		descZh:  "禁止 curl 管道到 shell: curl ... | sh/bash",
		descEn:  "Forbidden curl piped to shell: curl ... | sh/bash",
	},

	// 14. wget 管道到 shell
	{
		pattern: regexp.MustCompile(`\bwget\s+.*\|\s*(?:ba)?sh\b`),
		descZh:  "禁止 wget 管道到 shell: wget ... | sh/bash",
		descEn:  "Forbidden wget piped to shell: wget ... | sh/bash",
	},

	// 15. fdisk / parted 磁盘分区工具
	{
		pattern: regexp.MustCompile(`\b(?:fdisk|parted|gdisk|sfdisk)\b`),
		descZh:  "禁止磁盘分区操作: fdisk/parted/gdisk/sfdisk",
		descEn:  "Forbidden disk partition operation: fdisk/parted/gdisk/sfdisk",
	},

	// 16. chown 修改根路径文件所有者
	{
		pattern: regexp.MustCompile(`\bchown\s+(-R\s+)?[a-zA-Z0-9_\-\.]+:?[a-zA-Z0-9_\-\.]*\s+/`),
		descZh:  "禁止对根路径执行 chown 修改所有者",
		descEn:  "Forbidden chown on root path",
	},

	// 17. kill -9 强制终止进程
	{
		pattern: regexp.MustCompile(`\bkill\s+-9\b`),
		descZh:  "禁止 kill -9 强制终止进程",
		descEn:  "Forbidden kill -9 force terminate process",
	},

	// 18. 修改 crontab 计划任务
	{
		pattern: regexp.MustCompile(`\bcrontab\s+-`),
		descZh:  "禁止修改 crontab 计划任务",
		descEn:  "Forbidden crontab modification",
	},

	// 19. 访问 SSH 私钥及授权文件
	{
		pattern: regexp.MustCompile(`\.ssh/(?:id_rsa|id_ed25519|id_ecdsa|authorized_keys)`),
		descZh:  "禁止访问 SSH 私钥及授权文件",
		descEn:  "Forbidden access to SSH private keys and authorized keys",
	},

	// 20. 访问 .env 及云服务凭证文件
	{
		pattern: regexp.MustCompile(`\.env\b|credentials\.(?:json|yaml|yml)|\.aws/(?:credentials|config)`),
		descZh:  "禁止访问 .env 环境变量文件及云服务凭证文件",
		descEn:  "Forbidden access to .env files and cloud service credential files",
	},

	// 21. systemctl 管理系统服务
	{
		pattern: regexp.MustCompile(`\bsystemctl\s+(?:stop|disable|mask|enable|start|restart)\b`),
		descZh:  "禁止 systemctl 管理系统服务 (stop/disable/mask/enable/start/restart)",
		descEn:  "Forbidden systemctl service management (stop/disable/mask/enable/start/restart)",
	},

	// 22. 用户账户管理命令
	{
		pattern: regexp.MustCompile(`\b(?:usermod|useradd|userdel)\b`),
		descZh:  "禁止用户账户管理命令: usermod/useradd/userdel",
		descEn:  "Forbidden user account management: usermod/useradd/userdel",
	},

	// 23. passwd 修改密码
	{
		pattern: regexp.MustCompile(`\bpasswd\b`),
		descZh:  "禁止 passwd 修改用户密码",
		descEn:  "Forbidden passwd password modification",
	},

	// 25. mount / umount 挂载操作
	{
		pattern: regexp.MustCompile(`\b(?:mount|umount)\b`),
		descZh:  "禁止 mount/umount 挂载操作",
		descEn:  "Forbidden mount/umount operations",
	},

	// 26. history -c 清除命令历史
	{
		pattern: regexp.MustCompile(`\bhistory\s+-c\b`),
		descZh:  "禁止 history -c 清除命令历史",
		descEn:  "Forbidden history -c to clear command history",
	},

	// 27. 输出重定向覆盖 /etc 关键配置
	{
		pattern: regexp.MustCompile(`>\s*/etc/(?:hosts|hostname|resolv\.conf|fstab|sudoers)\b`),
		descZh:  "禁止重定向覆盖 /etc 下的关键配置文件",
		descEn:  "Forbidden redirection overwriting critical config files under /etc",
	},

	// 29. 写入 SSH authorized_keys
	{
		pattern: regexp.MustCompile(`(?:echo|cat|tee).*>>.*\.ssh/authorized_keys`),
		descZh:  "禁止写入 .ssh/authorized_keys 实现免密持久化",
		descEn:  "Forbidden writing to .ssh/authorized_keys for passwordless persistence",
	},

	// 30. Docker 特权模式或挂载宿主机根文件系统
	{
		pattern: regexp.MustCompile(`\bdocker\s+(?:exec|run)\s+.*--privileged\b|\bdocker\s+(?:exec|run)\s+.*-v\s+/:/host`),
		descZh:  "禁止 Docker 特权模式运行或挂载宿主机根文件系统",
		descEn:  "Forbidden Docker privileged mode or mounting host root filesystem",
	},

	// ─── Linux 系统管理 ───

	// 31. sysctl 内核参数修改
	{
		pattern: regexp.MustCompile(`\bsysctl\s+-w\b`),
		descZh:  "禁止 sysctl -w 修改内核参数",
		descEn:  "Forbidden sysctl kernel parameter modification",
	},

	// 32. modprobe / insmod / rmmod 内核模块管理
	{
		pattern: regexp.MustCompile(`\b(?:modprobe|insmod|rmmod|depmod)\b`),
		descZh:  "禁止加载/卸载内核模块: modprobe/insmod/rmmod/depmod",
		descEn:  "Forbidden kernel module management: modprobe/insmod/rmmod/depmod",
	},

	// 33. iptables / nft 防火墙规则完整操作
	{
		pattern: regexp.MustCompile(`\biptables\s+-[ADILRNX]\b|\bnft\s+(?:add|delete|insert|flush)\b`),
		descZh:  "禁止修改 iptables / nft 防火墙规则",
		descEn:  "Forbidden iptables / nft firewall rule modification",
	},

	// 34. firewalld / ufw 防火墙管理
	{
		pattern: regexp.MustCompile(`\b(?:firewall-cmd|ufw)\b`),
		descZh:  "禁止 firewalld / ufw 防火墙管理",
		descEn:  "Forbidden firewalld / ufw firewall management",
	},

	// 35. ip 命令操作网络接口和路由
	{
		pattern: regexp.MustCompile(`\bip\s+(?:addr\s+(?:add|del|flush)|route\s+(?:add|del|replace|flush)|link\s+(?:set|add|del))\b`),
		descZh:  "禁止 ip 命令修改网络接口和路由表",
		descEn:  "Forbidden ip command network interface and route modification",
	},

	// 36. chroot 切换根目录
	{
		pattern: regexp.MustCompile(`\bchroot\b`),
		descZh:  "禁止 chroot 切换根目录",
		descEn:  "Forbidden chroot root directory change",
	},

	// 37. nsenter 进入命名空间，可能造成容器逃逸
	{
		pattern: regexp.MustCompile(`\bnsenter\b`),
		descZh:  "禁止 nsenter 进入命名空间（容器逃逸风险）",
		descEn:  "Forbidden nsenter namespace entry (container escape risk)",
	},

	// 38. setcap / getcap 文件能力设置
	{
		pattern: regexp.MustCompile(`\bsetcap\b|\bgetcap\b`),
		descZh:  "禁止 setcap/getcap 操作文件能力",
		descEn:  "Forbidden setcap/getcap file capabilities manipulation",
	},

	// 39. strace / ltrace 进程跟踪（信息泄露）
	{
		pattern: regexp.MustCompile(`\b(?:strace|ltrace)\b`),
		descZh:  "禁止 strace/ltrace 进程跟踪",
		descEn:  "Forbidden strace/ltrace process tracing",
	},

	// 40. ldconfig 动态链接器缓存
	{
		pattern: regexp.MustCompile(`\bldconfig\b`),
		descZh:  "禁止 ldconfig 修改动态链接器配置",
		descEn:  "Forbidden ldconfig dynamic linker configuration",
	},

	// 41. SELinux 安全策略管理
	{
		pattern: regexp.MustCompile(`\b(?:setenforce|semanage|restorecon|chcon)\b`),
		descZh:  "禁止修改 SELinux 安全策略: setenforce/semanage/restorecon/chcon",
		descEn:  "Forbidden SELinux policy modification: setenforce/semanage/restorecon/chcon",
	},

	// 42. swap 交换空间管理
	{
		pattern: regexp.MustCompile(`\b(?:swapon|swapoff|mkswap)\b`),
		descZh:  "禁止 swap 交换空间操作: swapon/swapoff/mkswap",
		descEn:  "Forbidden swap space management: swapon/swapoff/mkswap",
	},

	// 43. LVM 逻辑卷管理
	{
		pattern: regexp.MustCompile(`\b(?:lvcreate|lvremove|lvextend|lvresize|vgcreate|vgremove|pvcreate|pvremove)\b`),
		descZh:  "禁止 LVM 逻辑卷管理操作",
		descEn:  "Forbidden LVM logical volume management",
	},

	// 44. unshare 命名空间操作
	{
		pattern: regexp.MustCompile(`\bunshare\b`),
		descZh:  "禁止 unshare 命名空间操作",
		descEn:  "Forbidden unshare namespace operation",
	},

	// 45. hostnamectl / timedatectl 系统标识修改
	{
		pattern: regexp.MustCompile(`\bhostnamectl\s+set\b|\btimedatectl\s+set\b`),
		descZh:  "禁止 hostnamectl/timedatectl 修改系统标识和时间",
		descEn:  "Forbidden hostnamectl/timedatectl system identity and time modification",
	},
}

// strictSafetyRules 严格安全审核规则，用于 MainAgent。
// 继承 baseSafetyRules 并追加 extraStrictRules。
var strictSafetyRules []safetyRule

func init() {
	strictSafetyRules = make([]safetyRule, len(baseSafetyRules)+len(extraStrictRules))
	copy(strictSafetyRules, baseSafetyRules)
	copy(strictSafetyRules[len(baseSafetyRules):], extraStrictRules)
}

// ValidateCommand 严格安全审核，用于 MainAgent。
func ValidateCommand(cmd string) error {
	return validateWithRules(cmd, strictSafetyRules)
}

// ValidateCommandForSystem 宽松安全审核，用于 SystemAgent。
// 仅使用基础规则，允许 skill 依赖安装等操作。
func ValidateCommandForSystem(cmd string) error {
	return validateWithRules(cmd, baseSafetyRules)
}

func validateWithRules(cmd string, rules []safetyRule) error {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		return nil
	}

	isZh := config.Get().GetLanguage() == "zh"
	prefix := getSafetyPrefix()
	lower := strings.ToLower(trimmed)
	for _, rule := range rules {
		if rule.pattern.MatchString(lower) {
			desc := rule.descEn
			if isZh {
				desc = rule.descZh
			}
			return fmt.Errorf("%s%s", prefix, desc)
		}
	}

	return nil
}
