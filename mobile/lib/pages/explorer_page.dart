import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/explorer_provider.dart';
import '../providers/settings_provider.dart';
import '../widgets/file_item_widget.dart';
import '../widgets/breadcrumbs_widget.dart';
import '../models.dart';
import 'preview_page.dart';
import 'settings_page.dart';
import 'waterfall_page.dart';

class ExplorerPage extends StatefulWidget {
  const ExplorerPage({Key? key}) : super(key: key);

  @override
  State<ExplorerPage> createState() => _ExplorerPageState();
}

class _ExplorerPageState extends State<ExplorerPage> {
  final SearchController _searchController = SearchController();

  @override
  Widget build(BuildContext context) {
    final explorer = context.watch<ExplorerProvider>();
    final settings = context.watch<SettingsProvider>();

    return PopScope(
      canPop: !explorer.canGoBack,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        if (explorer.canGoBack) {
          explorer.goBack();
        }
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Invisible Archive'),
          actions: [
            SearchAnchor(
              searchController: _searchController,
              builder: (context, controller) {
                return IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: () => controller.openView(),
                );
              },
              suggestionsBuilder: (context, controller) async {
                if (controller.text.isEmpty) return [];
                final results = await explorer.api.search(controller.text);
                return results.map((item) => ListTile(
                      leading: const Icon(Icons.history),
                      title: Text(item.name),
                      onTap: () {
                        controller.closeView(item.name);
                        explorer.performSearch(item.name);
                      },
                    ));
              },
            ),
            IconButton(
              icon: Icon(settings.isDarkMode ? Icons.light_mode : Icons.dark_mode),
              onPressed: () => settings.setDarkMode(!settings.isDarkMode),
            ),
            IconButton(
              icon: const Icon(Icons.auto_awesome_motion),
              tooltip: 'Waterfall View',
              onPressed: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => WaterfallPage(path: explorer.currentPath, api: explorer.api),
                ),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.settings),
              onPressed: () => Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const SettingsPage()),
              ),
            ),
          ],
        ),
        body: SafeArea(
          child: Column(
            children: [
              BreadcrumbsWidget(
                path: explorer.currentPath,
                onNavigate: (path) => explorer.navigateTo(path),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        '${explorer.items.length} items',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ),
                    PopupMenuButton<String>(
                      initialValue: explorer.sortMode,
                      onSelected: explorer.setSortMode,
                      icon: const Icon(Icons.sort, size: 20),
                      tooltip: 'Sort Mode',
                      itemBuilder: (context) => [
                        const PopupMenuItem(value: 'name', child: Text('Name')),
                        const PopupMenuItem(value: 'natural', child: Text('Natural')),
                        const PopupMenuItem(value: 'size', child: Text('Size')),
                        const PopupMenuItem(value: 'random', child: Text('Shuffle')),
                      ],
                    ),
                    if (explorer.sortMode != 'random')
                      IconButton(
                        icon: Icon(
                          explorer.isDescending ? Icons.arrow_downward : Icons.arrow_upward,
                          size: 18,
                        ),
                        onPressed: explorer.toggleSortOrder,
                        tooltip: 'Toggle Order',
                        visualDensity: VisualDensity.compact,
                      ),
                    const SizedBox(width: 8),
                    SegmentedButton<String>(
                      segments: const [
                        ButtonSegment(value: 'grid', icon: Icon(Icons.grid_view), tooltip: 'Grid'),
                        ButtonSegment(value: 'list', icon: Icon(Icons.view_list), tooltip: 'List'),
                        ButtonSegment(value: 'details', icon: Icon(Icons.list), tooltip: 'Details'),
                      ],
                      selected: {settings.layoutMode},
                      onSelectionChanged: (Set<String> selection) {
                        settings.setLayoutMode(selection.first);
                      },
                      showSelectedIcon: false,
                      style: const ButtonStyle(visualDensity: VisualDensity.compact),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              Expanded(
                child: RefreshIndicator(
                  onRefresh: () => explorer.fetchList(explorer.currentPath),
                  child: _buildContent(explorer, settings),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent(ExplorerProvider explorer, SettingsProvider settings) {
    if (explorer.isLoading && explorer.items.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (explorer.error != null && explorer.items.isEmpty) {
      final cs = Theme.of(context).colorScheme;
      return Center(
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error_outline, color: cs.error, size: 48),
              const SizedBox(height: 16),
              Text('Error: ${explorer.error}'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => explorer.fetchList(explorer.currentPath),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    if (explorer.items.isEmpty) {
      return const Center(
        child: SingleChildScrollView(
          physics: AlwaysScrollableScrollPhysics(),
          child: Text('No items found'),
        ),
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        if (settings.layoutMode == 'grid') {
          // Dynamic crossAxisCount based on width
          final int crossAxisCount = (constraints.maxWidth / 120).floor().clamp(3, 10);
          return GridView.builder(
            padding: const EdgeInsets.all(8),
            gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: crossAxisCount,
              crossAxisSpacing: 4,
              mainAxisSpacing: 4,
              childAspectRatio: 0.85,
            ),
            itemCount: explorer.items.length,
            itemBuilder: (context, index) {
              final item = explorer.items[index];
              return FileItemWidget(
                item: item,
                layout: 'grid',
                api: explorer.api,
                onTap: () => _handleItemTap(context, item, explorer),
              );
            },
          );
        } else {
          return ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: explorer.items.length,
            itemBuilder: (context, index) {
              final item = explorer.items[index];
              return FileItemWidget(
                item: item,
                layout: settings.layoutMode,
                api: explorer.api,
                onTap: () => _handleItemTap(context, item, explorer),
              );
            },
          );
        }
      },
    );
  }

  void _handleItemTap(BuildContext context, FileItem item, ExplorerProvider explorer) {
    if (item.canBrowse) {
      explorer.navigateTo(item.path);
    } else {
      final allImages = explorer.items.where((i) => i.canRender && !i.name.toLowerCase().endsWith('.pdf')).toList();
      final initialIndex = allImages.indexWhere((i) => i.path == item.path);

      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => PreviewPage(
            item: item,
            api: explorer.api,
            allImages: initialIndex != -1 ? allImages : null,
            initialIndex: initialIndex != -1 ? initialIndex : null,
          ),
        ),
      );
    }
  }
}
