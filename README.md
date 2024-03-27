# GenshinSymlinker

通过创建符号链接的方式来实现「原神/星铁」的「国服/国际服」双版本共存

## 使用方法

> 请注意，符号链接尽量不要跨分区使用，可能会造成游戏无法正常加载

前提是有下载好的游戏本体（任意服），以及对应服的换服包（v1.0.3 之后可为空目录，会自动下载换服包）。

我本地主要是国际服，国服是换服包+符号链接，以下说明以我本地为例。

国际服启动器位于：`D:\Genshin Impact`

国际服游戏根目录（包含 `GenshinImpact.exe` 文件的那个目录）：`D:\Genshin Impact\Genshin Impact game`

国服换服包放置于：`D:\Genshin Impact\YuanShen`

然后使用软件，分别输入国际服游戏根目录（本体） `D:\Genshin Impact\Genshin Impact game` 以及换服包目录 `D:\Genshin Impact\YuanShen`。

软件将自动比对两个文件夹，并在国服换服包所在目录下自动生成所需的符号链接，指向国际服的文件或者文件夹。

------

执行完毕后，国服可直接运行 `D:\Genshin Impact\YuanShen\YuanShen.exe` 启动。

之后切换只需要执行对应的 exe 文件即可，无需反复更换。

版本更新需要删除国服的文件夹 `YuanShen`，然后重新执行上述步骤一次。
