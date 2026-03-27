import 'dart:async';
import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../models.dart';
import '../api.dart';

class WaterfallItemWidget extends StatefulWidget {
  final FileItem item;
  final ApiService api;
  final VoidCallback onTap;

  const WaterfallItemWidget({
    Key? key,
    required this.item,
    required this.api,
    required this.onTap,
  }) : super(key: key);

  @override
  State<WaterfallItemWidget> createState() => _WaterfallItemWidgetState();
}

class _WaterfallItemWidgetState extends State<WaterfallItemWidget> {
  bool _showRaw = false;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    // Start loading the raw image after a small delay to avoid jank during scrolling
    _timer = Timer(const Duration(milliseconds: 300), () {
      if (mounted) {
        setState(() {
          _showRaw = true;
        });
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    // Explicitly evict the raw image from memory cache to save memory when out of view
    // This follows the "Swap & Unload" strategy
    final rawUrl = widget.api.getRawUrl(widget.item.path);
    PaintingBinding.instance.imageCache.evict(CachedNetworkImageProvider(rawUrl));
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final thumbUrl = widget.api.getThumbUrl(widget.item.path);
    final rawUrl = widget.api.getRawUrl(widget.item.path);
    final cs = Theme.of(context).colorScheme;
    final sh = Theme.of(context).shadowColor;
    return Card(
      margin: EdgeInsets.zero,
      elevation: 4,
      shadowColor: Theme.of(context).shadowColor.withOpacity(0.08),
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        onTap: widget.onTap,
        child: Stack(
          children: [
            // Tier 1: Thumbnail (Always loaded as base)
            CachedNetworkImage(
              imageUrl: thumbUrl,
              placeholder: (context, url) => AspectRatio(
                aspectRatio: 1,
                child: Container(
                  color: cs.surfaceVariant,
                  child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                ),
              ),
              errorWidget: (context, url, error) => AspectRatio(
                aspectRatio: 1,
                child: Center(child: Icon(Icons.broken_image_outlined, color: cs.onSurface)),
              ),
              fit: BoxFit.cover,
              width: double.infinity,
            ),

            // Tier 2: Original (Loaded and overlaid when widget is stable)
            if (_showRaw)
              CachedNetworkImage(
                imageUrl: rawUrl,
                placeholder: (context, url) => const SizedBox.shrink(),
                errorWidget: (context, url, error) => const SizedBox.shrink(),
                fit: BoxFit.cover,
                width: double.infinity,
                fadeInDuration: const Duration(milliseconds: 500),
              ),

            // Subtle gradient overlay for text readability
            Positioned.fill(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      Colors.transparent,
                      cs.onSurface.withOpacity(0.04),
                      cs.onSurface.withOpacity(0.32),
                    ],
                    stops: const [0.6, 0.8, 1.0],
                  ),
                ),
              ),
            ),

            // Item Info
            Positioned(
              bottom: 8,
              left: 8,
              right: 8,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    widget.item.name,
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                          color: cs.onSurface,
                          fontWeight: FontWeight.w700,
                          shadows: [Shadow(blurRadius: 2, color: sh.withOpacity(0.6))],
                        ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
