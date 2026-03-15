import 'package:flutter/material.dart';
import 'package:flutter_staggered_grid_view/flutter_staggered_grid_view.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../models.dart';
import '../api.dart';
import 'preview_page.dart';

class WaterfallPage extends StatefulWidget {
  final String path;
  final ApiService api;

  const WaterfallPage({
    Key? key,
    required this.path,
    required this.api,
  }) : super(key: key);

  @override
  State<WaterfallPage> createState() => _WaterfallPageState();
}

class _WaterfallPageState extends State<WaterfallPage> {
  final List<FileItem> _items = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  String? _error;
  final ScrollController _scrollController = ScrollController();
  final int _pageSize = 60;

  @override
  void initState() {
    super.initState();
    _loadInitialImages();
    _scrollController.addListener(_onScroll);
  }

  void _onScroll() {
    if (!_scrollController.hasClients) return;
    
    // If we are within 500 pixels of the bottom, load more
    if (_scrollController.position.pixels >= _scrollController.position.maxScrollExtent - 500) {
      if (!_isLoading && !_isLoadingMore && _hasMore) {
        _loadMoreImages();
      }
    }
  }

  Future<void> _loadInitialImages() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _items.clear();
      _error = null;
      _hasMore = true;
    });

    try {
      final items = await widget.api.fetchRandom(widget.path, limit: _pageSize);
      if (mounted) {
        setState(() {
          _items.addAll(items);
          _isLoading = false;
          if (items.length < _pageSize) {
            _hasMore = false;
          }
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMoreImages() async {
    if (!mounted || _isLoadingMore || !_hasMore) return;

    setState(() {
      _isLoadingMore = true;
    });

    try {
      final items = await widget.api.fetchRandom(widget.path, limit: _pageSize);
      if (mounted) {
        setState(() {
          // Avoid duplicates if random returns same images
          final existingPaths = _items.map((i) => i.path).toSet();
          final newItems = items.where((it) => !existingPaths.contains(it.path)).toList();
          
          if (newItems.isEmpty && items.isNotEmpty) {
            // If we got items but they are all duplicates, we might want to stop or retry once
            // For random discovery, getting some duplicates is expected, but if we get 0 new ones
            // after a few tries we should probably stop. For now just add what we got.
          }

          _items.addAll(newItems);
          _isLoadingMore = false;
          
          // Since it's RANDOM, we don't strictly know if there's more, 
          // but if we get fewer than we asked for, we've likely hit the total count.
          if (items.length < _pageSize) {
            _hasMore = false;
          }
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoadingMore = false;
          // Don't set global error for "load more" failure, maybe just a toast
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Error loading more: $e')),
          );
        });
      }
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Waterfall View'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadInitialImages,
          ),
        ],
      ),
      body: SafeArea(child: _buildBody()),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_error != null) {
      return Center(
        child: SingleChildScrollView(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, color: Colors.red, size: 48),
              const SizedBox(height: 16),
              Text('Error: $_error'),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _loadInitialImages,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    if (_items.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.image_not_supported_outlined, 
                 size: 64, 
                 color: Theme.of(context).colorScheme.outline),
            const SizedBox(height: 16),
            const Text('No images found recursive under this path.'),
          ],
        ),
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        final int crossAxisCount = (constraints.maxWidth / 180).floor().clamp(2, 8);
        return MasonryGridView.count(
          controller: _scrollController,
          padding: const EdgeInsets.all(4),
          crossAxisCount: crossAxisCount,
          mainAxisSpacing: 4,
          crossAxisSpacing: 4,
          itemCount: _items.length + (_hasMore ? 1 : 0),
          itemBuilder: (context, index) {
            if (index == _items.length) {
              return const Padding(
                padding: EdgeInsets.symmetric(vertical: 32),
                child: Center(child: CircularProgressIndicator()),
              );
            }

            final item = _items[index];
            return Card(
              margin: EdgeInsets.zero,
              elevation: 0,
              clipBehavior: Clip.antiAlias,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
              child: InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => PreviewPage(
                        item: item,
                        api: widget.api,
                        allImages: _items,
                        initialIndex: index,
                      ),
                    ),
                  );
                },
                child: CachedNetworkImage(
                  imageUrl: widget.api.getThumbUrl(item.path),
                  placeholder: (context, url) => AspectRatio(
                    aspectRatio: 1,
                    child: Container(
                      color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
                      child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                    ),
                  ),
                  errorWidget: (context, url, error) => AspectRatio(
                    aspectRatio: 1,
                    child: Container(
                      color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
                      child: const Icon(Icons.image_not_supported),
                    ),
                  ),
                  fit: BoxFit.cover,
                ),
              ),
            );
          },
        );
      },
    );
  }
}
