# GoPuppy Design Spec · 暖陶管家

## Direction

不是科技紫渐变，也不是医院冷白。GoPuppy 的视觉记忆点是**窗台午后**：陶土橙、苔藓绿、羊皮纸底，像一本被摸旧的宠物手帐。

## Palette

| Token | Hex | Use |
|---|---|---|
| `--paper` | `#F3E6D4` | 页面底 |
| `--card` | `#FFF8EE` | 卡片 |
| `--ink` | `#2A2118` | 正文 |
| `--clay` | `#C45C26` | 主操作 / 喂食 |
| `--moss` | `#3D6B4F` | 健康 / 成功 |
| `--gold` | `#E0A100` | 提醒 |
| `--rose` | `#B42318` | 异常 / 危险 |
| `--line` | `#E4D2B8` | 分割 |

## Typography

- Display：`Fraunces`（宠物名、大数字年龄）
- Body：`Noto Serif SC` + `Source Serif 4`
- 禁止 Inter / Roboto / 系统默认堆砌

## Components

- 圆角 18–24px，阴影用暖色 `rgba(90,50,20,.08)`
- 按钮：陶土实心 + 按下下沉 1px；禁用降低饱和
- 四态：loading 骨架（米色波纹）、empty 插画文案、error 可关闭 toast（5s + ×）、success 苔绿轻提示
- 原生 `alert/confirm` 禁止；统一 Modal
- Select 全局自定义箭头（折线 SVG）

## Breakpoints

375 / 768 / 1280。页面容器 `w-full`，不使用内容区 `max-w-*`（登录卡片与 Modal 例外）。
