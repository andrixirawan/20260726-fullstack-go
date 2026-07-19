"use client"

import { useState, useEffect, use } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import {
  ArrowLeft,
  Edit,
  Trash2,
  Globe,
  FileText,
  Eye,
  Calendar,
  Tag,
  Folder,
  Send,
  MessageCircle,
  Reply,
  MoreVertical,
} from "lucide-react"
import { toast } from "sonner"

import { blogApi, Post, CommentResponse, ApiError } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useAuthStore } from "@/store/auth.store"

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2)
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  })
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return "just now"
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

// Render simple markdown as plain HTML (basic inline + headings)
function renderMarkdown(md: string): string {
  return md
    .split("\n")
    .map((line) => {
      if (/^### /.test(line)) return `<h3>${line.slice(4)}</h3>`
      if (/^## /.test(line)) return `<h2>${line.slice(3)}</h2>`
      if (/^# /.test(line)) return `<h1>${line.slice(2)}</h1>`
      if (line === "") return "<br/>"
      line = line.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
      line = line.replace(/\*(.+?)\*/g, "<em>$1</em>")
      line = line.replace(/`(.+?)`/g, "<code>$1</code>")
      return `<p>${line}</p>`
    })
    .join("")
}

interface CommentItemProps {
  comment: CommentResponse
  postId: string
  depth?: number
  onReload: () => void
}

function CommentItem({ comment, postId, depth = 0, onReload }: CommentItemProps) {
  const { user } = useAuthStore()
  const [replying, setReplying] = useState(false)
  const [editing, setEditing] = useState(false)
  const [replyText, setReplyText] = useState("")
  const [editText, setEditText] = useState(comment.content)
  const [loading, setLoading] = useState(false)

  const isOwner = user?.id === comment.author.id

  const handleReply = async () => {
    if (!replyText.trim()) return
    setLoading(true)
    try {
      await blogApi.createComment(postId, { content: replyText, parent_id: comment.id })
      setReplyText("")
      setReplying(false)
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to post reply")
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = async () => {
    if (!editText.trim()) return
    setLoading(true)
    try {
      await blogApi.updateComment(comment.id, { content: editText })
      setEditing(false)
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to update comment")
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!confirm("Delete this comment?")) return
    try {
      await blogApi.deleteComment(comment.id)
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete comment")
    }
  }

  return (
    <div className={`${depth > 0 ? "ml-8 border-l-2 border-border pl-4" : ""} space-y-3`}>
      <div className="flex items-start gap-3">
        <Avatar className="h-7 w-7 shrink-0 mt-0.5">
          <AvatarImage src={comment.author.avatar_url} />
          <AvatarFallback className="text-[10px]">
            {initials(comment.author.full_name || comment.author.email)}
          </AvatarFallback>
        </Avatar>

        <div className="flex-1 space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">{comment.author.full_name}</span>
            <span className="text-xs text-muted-foreground">{timeAgo(comment.created_at)}</span>
            {comment.is_deleted && (
              <Badge variant="outline" className="text-xs py-0">
                deleted
              </Badge>
            )}
          </div>

          {editing ? (
            <div className="space-y-2">
              <Textarea
                id={`edit-comment-${comment.id}`}
                value={editText}
                onChange={(e) => setEditText(e.target.value)}
                rows={3}
                className="resize-none text-sm"
              />
              <div className="flex gap-2">
                <Button
                  id={`btn-save-comment-${comment.id}`}
                  size="sm"
                  onClick={handleEdit}
                  disabled={loading}
                >
                  Save
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => setEditing(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          ) : (
            <p className="text-sm text-foreground/80">{comment.content}</p>
          )}

          {!comment.is_deleted && !editing && (
            <div className="flex items-center gap-2">
              {user && (
                <button
                  id={`btn-reply-${comment.id}`}
                  onClick={() => setReplying(!replying)}
                  className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
                >
                  <Reply className="h-3 w-3" />
                  Reply
                </button>
              )}
              {isOwner && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <span className="flex items-center justify-center h-6 w-6 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer">
                      <MoreVertical className="h-3.5 w-3.5" />
                    </span>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start">
                    <DropdownMenuItem onClick={() => setEditing(true)}>Edit</DropdownMenuItem>
                    <DropdownMenuItem className="text-destructive" onClick={handleDelete}>
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          )}

          {replying && (
            <div className="space-y-2 pt-2">
              <Textarea
                id={`reply-input-${comment.id}`}
                placeholder="Write a reply..."
                value={replyText}
                onChange={(e) => setReplyText(e.target.value)}
                rows={2}
                className="resize-none text-sm"
              />
              <div className="flex gap-2">
                <Button
                  id={`btn-submit-reply-${comment.id}`}
                  size="sm"
                  onClick={handleReply}
                  disabled={loading || !replyText.trim()}
                  className="gap-1"
                >
                  <Send className="h-3 w-3" />
                  Reply
                </Button>
                <Button size="sm" variant="ghost" onClick={() => setReplying(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Nested replies */}
      {comment.replies?.length > 0 && (
        <div className="space-y-3">
          {comment.replies.map((reply) => (
            <CommentItem
              key={reply.id}
              comment={reply}
              postId={postId}
              depth={depth + 1}
              onReload={onReload}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default function PostDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const { user } = useAuthStore()

  const [post, setPost] = useState<Post | null>(null)
  const [comments, setComments] = useState<CommentResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [commentText, setCommentText] = useState("")
  const [submittingComment, setSubmittingComment] = useState(false)

  const loadPost = async () => {
    try {
      const [p, c] = await Promise.all([blogApi.getPostById(id), blogApi.listComments(id)])
      setPost(p)
      setComments(c)
    } catch {
      toast.error("Failed to load post")
      router.push("/blog")
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    loadPost()
  }, [id])

  const handleDelete = async () => {
    if (!confirm("Delete this post permanently?")) return
    try {
      await blogApi.deletePost(id)
      toast.success("Post deleted")
      router.push("/blog")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete post")
    }
  }

  const handleTogglePublish = async () => {
    if (!post) return
    try {
      const updated = await blogApi.togglePublish(id)
      setPost(updated)
      toast.success(updated.status === "published" ? "Post published!" : "Moved to draft")
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to update status")
    }
  }

  const handleCommentSubmit = async () => {
    if (!commentText.trim()) return
    setSubmittingComment(true)
    try {
      await blogApi.createComment(id, { content: commentText })
      setCommentText("")
      const updated = await blogApi.listComments(id)
      setComments(updated)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to post comment")
    } finally {
      setSubmittingComment(false)
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-3xl mx-auto space-y-6 animate-pulse">
        <div className="h-6 bg-muted rounded w-1/4" />
        <div className="h-10 bg-muted rounded w-3/4" />
        <div className="h-4 bg-muted rounded w-1/2" />
        <div className="h-64 bg-muted rounded" />
      </div>
    )
  }

  if (!post) return null

  const isOwner = user?.id === post.author.id
  const totalComments = comments.reduce(
    (acc, c) => acc + 1 + (c.replies?.length ?? 0),
    0,
  )

  return (
    <div className="max-w-3xl mx-auto space-y-8">
      {/* Nav */}
      <div className="flex items-center justify-between">
        <Link href="/blog">
          <Button id="btn-back-detail" variant="ghost" size="sm" className="gap-2">
            <ArrowLeft className="h-4 w-4" />
            Back to Blog
          </Button>
        </Link>
        {isOwner && (
          <div className="flex gap-2">
            <Button
              id="btn-toggle-publish"
              variant="outline"
              size="sm"
              onClick={handleTogglePublish}
              className="gap-1"
            >
              {post.status === "published" ? (
                <>
                  <FileText className="h-4 w-4" />
                  Move to Draft
                </>
              ) : (
                <>
                  <Globe className="h-4 w-4" />
                  Publish
                </>
              )}
            </Button>
            <Link href={`/blog/${id}/edit`}>
              <Button id="btn-edit-post" variant="outline" size="sm" className="gap-1">
                <Edit className="h-4 w-4" />
                Edit
              </Button>
            </Link>
            <Button
              id="btn-delete-post"
              variant="outline"
              size="sm"
              className="gap-1 text-destructive hover:text-destructive"
              onClick={handleDelete}
            >
              <Trash2 className="h-4 w-4" />
              Delete
            </Button>
          </div>
        )}
      </div>

      {/* Cover */}
      {post.cover_image && (
        <div className="rounded-2xl overflow-hidden h-56 bg-muted">
          <img src={post.cover_image} alt={post.title} className="w-full h-full object-cover" />
        </div>
      )}

      {/* Title + meta */}
      <div className="space-y-4">
        <div className="flex items-center gap-2 flex-wrap">
          <Badge
            variant="outline"
            className={`text-xs ${post.status === "published" ? "bg-emerald-500/15 text-emerald-600 border-emerald-500/30" : "bg-amber-500/15 text-amber-600 border-amber-500/30"}`}
          >
            {post.status === "published" ? (
              <Globe className="h-3 w-3 mr-1" />
            ) : (
              <FileText className="h-3 w-3 mr-1" />
            )}
            {post.status}
          </Badge>
          {post.category && (
            <span className="flex items-center gap-1 text-sm text-muted-foreground">
              <Folder className="h-3.5 w-3.5" />
              {post.category.name}
            </span>
          )}
        </div>

        <h1 className="text-3xl font-bold tracking-tight leading-tight">{post.title}</h1>

        {post.excerpt && (
          <p className="text-lg text-muted-foreground">{post.excerpt}</p>
        )}

        <div className="flex items-center gap-5 text-sm text-muted-foreground flex-wrap">
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <AvatarImage src={post.author.avatar_url} />
              <AvatarFallback className="text-[10px]">
                {initials(post.author.full_name || post.author.email)}
              </AvatarFallback>
            </Avatar>
            <span>{post.author.full_name}</span>
          </div>
          {post.published_at && (
            <span className="flex items-center gap-1">
              <Calendar className="h-3.5 w-3.5" />
              {formatDate(post.published_at)}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Eye className="h-3.5 w-3.5" />
            {post.view_count} views
          </span>
        </div>

        {post.tags.length > 0 && (
          <div className="flex items-center gap-2 flex-wrap">
            <Tag className="h-3.5 w-3.5 text-muted-foreground" />
            {post.tags.map((t) => (
              <span
                key={t.id}
                className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full"
              >
                {t.name}
              </span>
            ))}
          </div>
        )}
      </div>

      <Separator />

      {/* Content */}
      <div
        className="prose prose-neutral dark:prose-invert max-w-none text-sm leading-7 space-y-3"
        dangerouslySetInnerHTML={{ __html: renderMarkdown(post.content) }}
      />

      <Separator />

      {/* Comments */}
      <div className="space-y-6">
        <h2 className="text-xl font-semibold flex items-center gap-2">
          <MessageCircle className="h-5 w-5" />
          Comments ({totalComments})
        </h2>

        {/* Add comment */}
        {user ? (
          <div className="flex gap-3">
            <Avatar className="h-8 w-8 shrink-0 mt-0.5">
              <AvatarImage src={user.avatar_url} />
              <AvatarFallback className="text-xs">
                {initials(user.full_name || user.email)}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-2">
              <Textarea
                id="input-new-comment"
                placeholder="Share your thoughts..."
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
                rows={3}
                className="resize-none"
              />
              <Button
                id="btn-submit-comment"
                size="sm"
                onClick={handleCommentSubmit}
                disabled={submittingComment || !commentText.trim()}
                className="gap-2"
              >
                <Send className="h-3.5 w-3.5" />
                Post Comment
              </Button>
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            <Link href="/login" className="text-primary underline underline-offset-4">
              Sign in
            </Link>{" "}
            to leave a comment
          </p>
        )}

        {/* Comment list */}
        {comments.length > 0 ? (
          <div className="space-y-6">
            {comments.map((c) => (
              <CommentItem
                key={c.id}
                comment={c}
                postId={id}
                onReload={() => blogApi.listComments(id).then(setComments)}
              />
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground text-center py-6">
            No comments yet. Be the first to comment!
          </p>
        )}
      </div>
    </div>
  )
}
