# GenshinSymlinker

**English**:  
GenshinSymlinker allows users to run both Chinese and Global versions of *Genshin Impact*, *Honkai: Star Rail*, and *Zenless Zone Zero* on the same device by using symbolic links to share game files, reducing storage usage.  

**中文**:  
GenshinSymlinker 允许用户在同一设备上运行《原神》、《星铁》、《绝区零》的国服与国际服，通过符号链接共享游戏文件，减少存储空间占用。

## Notes / 注意事项

**English**:  
- Version 1.0.7: Added support for chunk downloads (since *Genshin Impact* 5.6 no longer provides zip packages).  
- Version 1.0.6: Added support for *Zenless Zone Zero* (since ZZZ 1.1 provided `res_list_url`).  
- Version 1.0.5: Uses the new launcher API; old launcher APIs are deprecated and cannot check for updates.  
- The `patchdown` directory contains a tool for direct downloading of version-switch packages. Compile it yourself if needed.  

**中文**:  
- 1.0.7 版本：增加 Chunk 下载支持（因《原神》5.6 不再提供 zip 包）。  
- 1.0.6 版本：添加《绝区零》支持（因 ZZZ 1.1 提供了 `res_list_url`）。  
- 1.0.5 版本：使用新启动器接口，旧版启动器接口已弃用，无法正常检查更新。  
- `patchdown` 目录下为直接下载换服包的工具，如有需求可自行编译。

## Usage / 使用方法

**English**:  
> **Note**: Avoid creating symbolic links across different disk partitions, as this may cause the game to fail to load.  

**Prerequisites**:  
- A downloaded game client (either Chinese or Global version).  
- A version-switch package for the other server (since v1.0.3, an empty directory is sufficient, as the tool can auto-download the package).  

**Example (using Global version as primary and Chinese version via symbolic links)**:  
- Global launcher location: `D:\Genshin Impact`  
- Global game root directory (containing `GenshinImpact.exe`): `D:\Genshin Impact\Genshin Impact game`  
- Chinese version-switch package location: `D:\Genshin Impact\YuanShen`  

**Steps**:  
1. Run the tool and input:  
   - Global game root directory: `D:\Genshin Impact\Genshin Impact game`  
   - Chinese version-switch package directory: `D:\Genshin Impact\YuanShen`  
2. The tool will compare the folders and create necessary symbolic links in the Chinese version-switch directory, pointing to the Global version’s files or folders.  
3. After completion, launch the Chinese version by running `D:\Genshin Impact\YuanShen\YuanShen.exe`.  

**Switching**:  
Simply run the corresponding `.exe` file for the desired version. No need to repeatedly swap files.  

**Updating**:  
Delete the Chinese version folder (`YuanShen`), redownload the version-switch package, and repeat the above steps.  

**中文**:  
> **注意**：符号链接尽量不要跨分区使用，可能会导致游戏无法正常加载。  

**前提**：  
- 已下载游戏本体（任意服）。  
- 对应服的换服包（v1.0.3 之后可为空目录，工具会自动下载换服包）。  

**示例（以国际服为主，国服通过换服包+符号链接）**：  
- 国际服启动器位置：`D:\Genshin Impact`  
- 国际服游戏根目录（包含 `GenshinImpact.exe` 的目录）：`D:\Genshin Impact\Genshin Impact game`  
- 国服换服包位置：`D:\Genshin Impact\YuanShen`  

**步骤**：  
1. 运行软件，分别输入：  
   - 国际服游戏根目录：`D:\Genshin Impact\Genshin Impact game`  
   - 国服换服包目录：`D:\Genshin Impact\YuanShen`  
2. 软件将自动比对两个文件夹，并在国服换服包目录下生成所需的符号链接，指向国际服的文件或文件夹。  
3. 执行完毕后，直接运行 `D:\Genshin Impact\YuanShen\YuanShen.exe` 启动国服。  

**切换**：  
只需运行对应版本的 `.exe` 文件即可，无需反复更换文件。  

**更新**：  
删除国服文件夹（`YuanShen`），重新下载换服包，然后重复上述步骤。
