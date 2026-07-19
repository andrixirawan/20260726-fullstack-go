"use client"

import { useState, useEffect, useCallback } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import {
  PenLine,
  Search,
  Tag,
  Folder,
  Eye,
  Clock,
  Globe,
  FileText,
  Trash2,
  Edit,
  MoreVertical,
  ChevronLeft,
  ChevronRight,
  BookOpen,
  Settings,
} from "lucide-react"
import { toast } from "sonner"

import { ManageMetadataDialog } from "@/components/blog/manage-metadata-dialog"

import { blogApi, PostListItem, Category, Tag as TagType, PostStatus } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useAuthStore } from "@/store/auth.store"
import { ApiError } from "@/lib/api"

const statusColors: Record<PostStatus, string> = {
  published: "bg-emerald-500/15 text-emerald-600 border-emerald-500/30",
  draft: "bg-amber-500/15 text-amber-600 border-amber-500/30",
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return "just now"
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export default function BlogPage() {
  const router = useRouter()
  const { user } = useAuthStore()

  const [posts, setPosts] = useState<PostListItem[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [tags, setTags] = useState<TagType[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(0)
  const [isLoading, setIsLoading] = useState(true)

  const [search, setSearch] = useState("")
  const [debouncedSearch, setDebouncedSearch] = useState("")
  const [status, setStatus] = useState<PostStatus | "">("")
  const [categoryId, setCategoryId] = useState("")
  const [tagId, setTagId] = useState("")
  const [page, setPage] = useState(1)
  const [manageOpen, setManageOpen] = useState(false)

  const searchParams = useSearchParams()

  // Sync URL search params to filter state
  useEffect(() => {
    const cat = searchParams.get("category_id") || ""
    const tag = searchParams.get("tag_id") || ""
    if (cat) setCategoryId(cat)
    if (tag) setTagId(tag)
  }, [searchParams])

  // Debounce search
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search), 400)
    return () => clearTimeout(t)
  }, [search])

  // Reset page when filters change
  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, status, categoryId, tagId])

  const loadPosts = useCallback(async () => {
    setIsLoading(true)
    try {
      const res = await blogApi.listPosts({
        page,
        page_size: 9,
        search: debouncedSearch,
        status: status || undefined,
        category_id: categoryId || undefined,
        tag_id: tagId || undefined,
      })
      setPosts(res.data ?? [])
      setTotal(res.total)
      setTotalPages(res.total_pages)
    } catch {
      toast.error("Failed to load posts")
    } finally {
      setIsLoading(false)
    }
  }, [page, debouncedSearch, status, categoryId, tagId])

  useEffect(() => {
    loadPosts()
  }, [loadPosts])

  const loadMetadata = useCallback(() => {
    blogApi.listCategories().then(setCategories).catch(() => {})
    blogApi.listTags().then(setTags).catch(() => {})
  }, [])

  useEffect(() => {
    loadMetadata()
  }, [loadMetadata])

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.preventDefault()
    if (!confirm("Delete this post permanently?")) return
    try {
      await blogApi.deletePost(id)
      toast.success("Post deleted")
      loadPosts()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete post")
    }
  }

  const handleTogglePublish = async (id: string, e: React.MouseEvent) => {
    e.preventDefault()
    try {
      const updated = await blogApi.togglePublish(id)
      toast.success(updated.status === "published" ? "Post published!" : "Post moved to draft")
      loadPosts()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to update status")
    }
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Blog</h1>
          <p className="text-muted-foreground text-sm mt-0.5">
            {total} post{total !== 1 ? "s" : ""} total
          </p>
        </div>
        <div className="flex gap-2">
          {user && (
            <Button
              id="btn-manage-metadata"
              variant="outline"
              onClick={() => setManageOpen(true)}
              className="gap-2"
            >
              <Settings className="h-4 w-4" />
              Manage Categories/Tags
            </Button>
          )}
          <Link href="/blog/new">
            <Button id="btn-new-post" className="gap-2">
              <PenLine className="h-4 w-4" />
              New Post
            </Button>
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            id="input-search-posts"
            placeholder="Search posts..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={status} onValueChange={(v) => setStatus(v as PostStatus | "")}>
          <SelectTrigger id="select-status" className="w-36">
            <SelectValue placeholder="All statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All statuses</SelectItem>
            <SelectItem value="published">Published</SelectItem>
            <SelectItem value="draft">Draft</SelectItem>
          </SelectContent>
        </Select>

        <Select value={categoryId} onValueChange={(val) => setCategoryId(val ?? "")}>
          <SelectTrigger id="select-category" className="w-40">
            <SelectValue placeholder="All categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All categories</SelectItem>
            {categories.map((c) => (
              <SelectItem key={c.id} value={c.id}>
                {c.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={tagId} onValueChange={(val) => setTagId(val ?? "")}>
          <SelectTrigger id="select-tag" className="w-36">
            <SelectValue placeholder="All tags" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All tags</SelectItem>
            {tags.map((t) => (
              <SelectItem key={t.id} value={t.id}>
                {t.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Posts grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="border rounded-xl p-5 space-y-3 animate-pulse bg-card">
              <div className="h-3 bg-muted rounded w-1/3" />
              <div className="h-5 bg-muted rounded w-3/4" />
              <div className="h-3 bg-muted rounded w-full" />
              <div className="h-3 bg-muted rounded w-2/3" />
            </div>
          ))}
        </div>
      ) : posts.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <BookOpen className="h-12 w-12 text-muted-foreground/40 mb-4" />
          <p className="text-lg font-medium text-muted-foreground">No posts found</p>
          <p className="text-sm text-muted-foreground/70 mt-1">
            {search || status || categoryId || tagId
              ? "Try adjusting your filters"
              : "Create your first post to get started"}
          </p>
          {!search && !status && !categoryId && !tagId && (
            <Link href="/blog/new">
              <Button className="mt-4 gap-2" variant="outline">
                <PenLine className="h-4 w-4" />
                Write a post
              </Button>
            </Link>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {posts.map((post) => {
            const isOwner = user?.id === post.author.id
            return (
              <Link key={post.id} href={`/blog/${post.id}`} className="group block">
                <article className="border rounded-xl p-5 bg-card hover:border-primary/40 hover:shadow-md transition-all duration-200 h-full flex flex-col gap-3">
                  {/* Cover */}
                  {post.cover_image && (
                    <div className="rounded-lg overflow-hidden h-36 bg-muted">
                      <img
                        src={post.cover_image}
                        alt={post.title}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    </div>
                  )}

                  {/* Meta: category + status */}
                  <div className="flex items-center gap-2 flex-wrap">
                    {post.category && (
                      <button
                        type="button"
                        onClick={(e) => {
                          e.preventDefault()
                          e.stopPropagation()
                          setCategoryId(post.category?.id || "")
                        }}
                        className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                      >
                        <Folder className="h-3 w-3" />
                        {post.category.name}
                      </button>
                    )}
                    <Badge
                      variant="outline"
                      className={`text-xs px-2 py-0 ${statusColors[post.status]}`}
                    >
                      {post.status === "published" ? (
                        <Globe className="h-3 w-3 mr-1" />
                      ) : (
                        <FileText className="h-3 w-3 mr-1" />
                      )}
                      {post.status}
                    </Badge>
                  </div>

                  {/* Title + excerpt */}
                  <div className="flex-1 space-y-1">
                    <h2 className="font-semibold text-base leading-snug group-hover:text-primary transition-colors line-clamp-2">
                      {post.title}
                    </h2>
                    {post.excerpt && (
                      <p className="text-sm text-muted-foreground line-clamp-2">{post.excerpt}</p>
                    )}
                  </div>

                  {/* Tags */}
                  {post.tags.length > 0 && (
                    <div className="flex items-center gap-1 flex-wrap">
                      <Tag className="h-3 w-3 text-muted-foreground" />
                      {post.tags.slice(0, 3).map((t) => (
                        <button
                          key={t.id}
                          type="button"
                          onClick={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            setTagId(t.id)
                          }}
                          className="text-xs bg-primary/10 text-primary hover:bg-primary/20 px-1.5 py-0.5 rounded-full transition-colors cursor-pointer"
                        >
                          {t.name}
                        </button>
                      ))}
                      {post.tags.length > 3 && (
                        <span className="text-xs text-muted-foreground">
                          +{post.tags.length - 3}
                        </span>
                      )}
                    </div>
                  )}

                  {/* Footer: author + stats + actions */}
                  <div className="flex items-center justify-between pt-1 border-t border-border/50">
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Eye className="h-3 w-3" />
                        {post.view_count}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {timeAgo(post.updated_at)}
                      </span>
                    </div>
                    {isOwner && (
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity"
                              id={`btn-post-menu-${post.id}`}
                              onClick={(e) => e.preventDefault()}
                            >
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          }
                        />
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.preventDefault()
                              router.push(`/blog/${post.id}/edit`)
                            }}
                          >
                            <Edit className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={(e) => handleTogglePublish(post.id, e)}>
                            {post.status === "published" ? (
                              <>
                                <FileText className="h-4 w-4 mr-2" />
                                Move to Draft
                              </>
                            ) : (
                              <>
                                <Globe className="h-4 w-4 mr-2" />
                                Publish
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={(e) => handleDelete(post.id, e)}
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </div>
                </article>
              </Link>
            )
          })}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-4">
          <Button
            id="btn-prev-page"
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="gap-1"
          >
            <ChevronLeft className="h-4 w-4" />
            Prev
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button
            id="btn-next-page"
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
            className="gap-1"
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}
      {/* Manage metadata dialog */}
      <ManageMetadataDialog
        open={manageOpen}
        onOpenChange={setManageOpen}
        onReload={loadMetadata}
      />
    </div>
  )
}
