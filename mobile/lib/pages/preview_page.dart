import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:video_player/video_player.dart';
import 'package:chewie/chewie.dart';
import 'package:photo_view/photo_view.dart';
import 'package:photo_view/photo_view_gallery.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../models.dart';
import '../api.dart';

class PreviewPage extends StatefulWidget {
  final FileItem item;
  final ApiService api;
  final List<FileItem>? allImages;
  final int? initialIndex;
  final Future<void> Function()? onLoadMore;

  const PreviewPage({
    Key? key,
    required this.item,
    required this.api,
    this.allImages,
    this.initialIndex,
    this.onLoadMore,
  }) : super(key: key);

  @override
  State<PreviewPage> createState() => _PreviewPageState();
}

class _PreviewPageState extends State<PreviewPage> {
  late FileItem _currentItem;
  late PageController _pageController;
  VideoPlayerController? _videoController;
  ChewieController? _chewieController;
  String? _textContent;
  bool _isLoadingText = false;
  bool _isFetchingMore = false;
  late int _currentIndex;

  @override
  void initState() {
    super.initState();
    _currentItem = widget.item;
    _currentIndex = widget.initialIndex ?? 0;
    _pageController = PageController(initialPage: _currentIndex);
    _initCurrentItem();
    
    // Initial preloading
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _precacheImages(_currentIndex);
      _checkLoadMore(_currentIndex);
    });
  }

  void _initCurrentItem() {
    if (_currentItem.canStream) {
      _initVideo();
    } else if (_currentItem.canEdit) {
      _loadText();
    }
  }

  void _checkLoadMore(int index) async {
    if (widget.allImages == null || widget.onLoadMore == null || _isFetchingMore) return;

    // If we are within 10 images of the end, load more
    if (index >= widget.allImages!.length - 10) {
      _isFetchingMore = true;
      try {
        await widget.onLoadMore!();
        if (mounted) {
          setState(() {
            // Rebuild to update itemCount in PhotoViewGallery
          });
        }
      } finally {
        _isFetchingMore = false;
      }
    }
  }

  void _precacheImages(int index) {
    if (widget.allImages == null || widget.allImages!.isEmpty) return;

    // Preload next image
    if (index + 1 < widget.allImages!.length) {
      final nextItem = widget.allImages![index + 1];
      precacheImage(
        CachedNetworkImageProvider(widget.api.getRawUrl(nextItem.path)),
        context,
      );
    }

    // Preload previous image
    if (index - 1 >= 0) {
      final prevItem = widget.allImages![index - 1];
      precacheImage(
        CachedNetworkImageProvider(widget.api.getRawUrl(prevItem.path)),
        context,
      );
    }
  }

  void _initVideo() {
    _chewieController?.dispose();
    _videoController?.dispose();
    _videoController = VideoPlayerController.networkUrl(Uri.parse(widget.api.getRawUrl(_currentItem.path)))
      ..initialize().then((_) {
        if (mounted) {
          setState(() {
            _chewieController = ChewieController(
              videoPlayerController: _videoController!,
              autoPlay: true,
              looping: false,
              aspectRatio: _videoController!.value.aspectRatio,
              showControls: true,
              materialProgressColors: ChewieProgressColors(
                playedColor: Colors.blue,
                handleColor: Colors.blue,
                backgroundColor: Colors.white24,
                bufferedColor: Colors.white54,
              ),
              placeholder: Container(color: Colors.black),
              autoInitialize: true,
            );
          });
        }
      });
  }

  void _loadText() async {
    setState(() => _isLoadingText = true);
    try {
      final content = await widget.api.fetchText(_currentItem.path);
      if (mounted) setState(() => _textContent = content);
    } catch (e) {
      if (mounted) setState(() => _textContent = 'Error loading text: $e');
    } finally {
      if (mounted) setState(() => _isLoadingText = false);
    }
  }

  @override
  void dispose() {
    _chewieController?.dispose();
    _videoController?.dispose();
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final bool isImage = _currentItem.canRender && !_currentItem.name.toLowerCase().endsWith('.pdf');
    final bool hasGallery = widget.allImages != null && widget.allImages!.isNotEmpty && isImage;

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        title: Text(_currentItem.name),
        backgroundColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        foregroundColor: Colors.white,
        centerTitle: true,
        actions: [
          IconButton(
            icon: const Icon(Icons.download),
            onPressed: () {
              // TODO: Implement download
            },
          ),
        ],
      ),
      extendBodyBehindAppBar: true,
      body: hasGallery ? _buildGallery() : _buildSingleView(),
    );
  }

  Widget _buildGallery() {
    return Shortcuts(
      shortcuts: <ShortcutActivator, Intent>{
        const SingleActivator(LogicalKeyboardKey.arrowLeft): const ScrollIntent(direction: AxisDirection.left),
        const SingleActivator(LogicalKeyboardKey.arrowRight): const ScrollIntent(direction: AxisDirection.right),
      },
      child: Actions(
        actions: <Type, Action<Intent>>{
          ScrollIntent: CallbackAction<ScrollIntent>(onInvoke: (intent) {
            if (intent.direction == AxisDirection.left && _currentIndex > 0) {
              _pageController.previousPage(
                duration: const Duration(milliseconds: 300),
                curve: Curves.easeInOut,
              );
            } else if (intent.direction == AxisDirection.right && _currentIndex < widget.allImages!.length - 1) {
              _pageController.nextPage(
                duration: const Duration(milliseconds: 300),
                curve: Curves.easeInOut,
              );
            }
            return null;
          }),
        },
        child: Focus(
          autofocus: true,
          child: PhotoViewGallery.builder(
            scrollPhysics: const BouncingScrollPhysics(),
            builder: (BuildContext context, int index) {
              final item = widget.allImages![index];
              return PhotoViewGalleryPageOptions(
                imageProvider: CachedNetworkImageProvider(widget.api.getRawUrl(item.path)),
                initialScale: PhotoViewComputedScale.contained,
                minScale: PhotoViewComputedScale.contained * 0.8,
                maxScale: PhotoViewComputedScale.covered * 2,
                heroAttributes: PhotoViewHeroAttributes(tag: item.path),
              );
            },
            itemCount: widget.allImages!.length,
            loadingBuilder: (context, event) => const Center(child: CircularProgressIndicator()),
            backgroundDecoration: const BoxDecoration(color: Colors.black),
            pageController: _pageController,
            onPageChanged: (index) {
              setState(() {
                _currentIndex = index;
                _currentItem = widget.allImages![index];
              });
              _precacheImages(index);
              _checkLoadMore(index);
            },
          ),
        ),
      ),
    );
  }

  Widget _buildSingleView() {
    if (_currentItem.canRender && !_currentItem.name.toLowerCase().endsWith('.pdf')) {
      return _buildImagePreview();
    } else if (_currentItem.canStream) {
      return _buildVideoPreview();
    } else if (_currentItem.canEdit) {
      return _buildTextPreview();
    } else {
      return _buildFallbackPreview();
    }
  }

  Widget _buildImagePreview() {
    return PhotoView(
      imageProvider: CachedNetworkImageProvider(widget.api.getRawUrl(_currentItem.path)),
      loadingBuilder: (context, event) => const Center(child: CircularProgressIndicator()),
      errorBuilder: (context, error, stackTrace) => const Center(
        child: Text('Failed to load image', style: TextStyle(color: Colors.white)),
      ),
    );
  }

  Widget _buildVideoPreview() {
    if (_chewieController == null || !_videoController!.value.isInitialized) {
      return const Center(child: CircularProgressIndicator());
    }
    return Center(
      child: Chewie(controller: _chewieController!),
    );
  }

  Widget _buildTextPreview() {
    if (_isLoadingText) {
      return const Center(child: CircularProgressIndicator());
    }
    return Container(
      color: Theme.of(context).colorScheme.surface,
      child: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: SelectableText(
            _textContent ?? '',
            style: TextStyle(
              color: Theme.of(context).colorScheme.onSurface,
              fontFamily: 'monospace',
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFallbackPreview() {
    return Container(
      color: Theme.of(context).colorScheme.surface,
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, 
                 color: Theme.of(context).colorScheme.error, 
                 size: 64),
            const SizedBox(height: 16),
            const Text('No preview available for this file type.'),
            const SizedBox(height: 24),
            FilledButton.tonal(
              onPressed: () {
                // TODO: Launch URL in browser
              },
              child: const Text('Open in Browser'),
            ),
          ],
        ),
      ),
    );
  }
}
