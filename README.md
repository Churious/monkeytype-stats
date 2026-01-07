
# MonkeyType Stats Viewer

Display your Monkeytype stats dynamically on your GitHub Profile. </br>
Supports all themes, modes, and custom configurations.

## Demo
`/api?user=Churious&theme=blueberry_dark&mode=time&length=30` </br>
![My Stats](https://monkeytype-stats.vercel.app/api?user=Churious&theme=blueberry_dark&mode=time&length=30)
## Deployment

Deploy your own instance using your favorite provider. **Vercel is recommended** for the best compatibility with Go.

| Provider | Deploy Link | Free Tier |
| :--- | :--- | :--- |
| **Vercel** | [![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/YOUR_GITHUB_USERNAME/monkeytype-stats) | ✅ (Recommended) |
| **Netlify** | [![Deploy to Netlify](https://www.netlify.com/img/deploy/button.svg)](https://app.netlify.com/start/deploy?repository=https://github.com/YOUR_GITHUB_USERNAME/monkeytype-stats) | ✅ |
| **Railway** | [![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template?template=https://github.com/YOUR_GITHUB_USERNAME/monkeytype-stats) | ⚠️ (Trial Only) |

## Configuration

You can customize your card by adding query parameters to the URL.

| Parameter | Description | Default | Example |
| :--- | :--- | :--- | :--- |
| `username` | **(Required)** Your Monkeytype username | - | `?username=MiDeco` |
| `theme` | Name of the theme ([supports ALL themes](https://github.com/monkeytypegame/monkeytype/tree/master/frontend/static/themes)) | `dark` | `?theme=serika_dark` |
| `mode` | Typing mode (`time` or `words`) | `time` | `?mode=words` |
| `length` | Test duration options:<br>• **time**: `15`, `30`, `60`, `120`<br>• **words**: `10`, `25`, `50`, `100` | `60` | `?length=25` |

> [!NOTE]
> Theme names are automatically fetched from Monkeytype. </br>
> Spaces in theme names should be replaced with underscores (e.g., `modern dolch` -> `modern_dolch`).


## FAQ

#### Q: My stats are not updating instantly!

A: GitHub caches images to improve performance. The image usually refreshes within 10-15 minutes. </br>
If you want to force an update, clear your browser cache or wait a moment.

#### Q: The theme looks different from Monkeytype.

A: Ensure the theme name is correct. Spaces must be replaced with underscores (e.g., `modern dolch` -> `modern_dolch`).
## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=churious/monkeytype-stats&type=Date)](https://star-history.com/#churious/monkeytype-stats&Date)