"use client"

import { useState, useEffect, use } from "react"
import { useRouter } from "next/navigation"
import { ArrowLeft, Save, Send, X, Upload, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { blogApi, Post, Category, Tag, PostStatus, ApiError, API_BASE_URL } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useAuthStore } from "@/store/auth.store"

export default function EditPostPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params)
  const router = useRouter()
  const { user } = useAuthStore()

  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [categories, setCategories] = useState<Category[]>([])
  const [tags, setTags] = useState<Tag[]>([])

  const [title, setTitle] = useState("")
  const [excerpt, setExcerpt] = useState("")
  const [content, setContent] = useState("")
  const [coverImage, setCoverImage] = useState("")
  const [isUploadingCover, setIsUploadingCover] = useState(false)
  const [categoryId, setCategoryId] = useState("")
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([])
  const [currentStatus, setCurrentStatus] = useState<PostStatus>("draft")

  const handleCoverUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setIsUploadingCover(true)
    try {
      const result = await blogApi.uploadImage(file)
      const coverUrl = `${API_BASE_URL}${result.url}`
      setCoverImage(coverUrl)
      toast.success("Cover image uploaded successfully!")
    } catch {
      toast.error("Failed to upload cover image")
    } finally {
      setIsUploadingCover(false)
    }
  }

  useEffect(() => {
    const load = async () => {
      try {
        const [post, cats, allTags] = await Promise.all([
          blogApi.getPostById(id),
          blogApi.listCategories(),
          blogApi.listTags(),
        ])

        // Only the author can edit
        if (post.author.id !== user?.id) {
          toast.error("You are not allowed to edit this post")
          router.push(`/blog/${id}`)
          return
        }

        setTitle(post.title)
        setExcerpt(post.excerpt)
        setContent(post.content)
        setCoverImage(post.cover_image)
        setCategoryId(post.category?.id ?? "")
        setSelectedTagIds(post.tags.map((t) => t.id))
        setCurrentStatus(post.status)
        setCategories(cats)
        setTags(allTags)
      } catch {
        toast.error("Failed to load post")
        router.push("/blog")
      } finally {
        setIsLoading(false)
      }
    }
    if (user) load()
  }, [id, user])

  const toggleTag = (tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((t) => t !== tagId) : [...prev, tagId],
    )
  }

  const handleSubmit = async (newStatus?: PostStatus) => {
    if (!title.trim()) {
      toast.error("Title is required")
      return
    }
    if (!content.trim()) {
      toast.error("Content is required")
      return
    }

    setIsSubmitting(true)
    try {
      await blogApi.updatePost(id, {
        title,
        excerpt,
        content,
        cover_image: coverImage,
        category_id: categoryId || undefined,
        tag_ids: selectedTagIds,
        status: newStatus,
      })
      toast.success("Post updated successfully!")
      router.push(`/blog/${id}`)
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to update post")
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto space-y-6 animate-pulse">
        <div className="h-8 bg-muted rounded w-1/4" />
        <div className="h-12 bg-muted rounded" />
        <div className="h-48 bg-muted rounded" />
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          id="btn-back-edit"
          variant="ghost"
          size="icon"
          onClick={() => router.back()}
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Edit Post</h1>
          <p className="text-sm text-muted-foreground">Update your blog post</p>
        </div>
        <div className="ml-auto flex gap-2">
          {currentStatus === "published" ? (
            <Button
              id="btn-update-draft"
              variant="outline"
              onClick={() => handleSubmit("draft")}
              disabled={isSubmitting}
              className="gap-2"
            >
              <Save className="h-4 w-4" />
              Save as Draft
            </Button>
          ) : (
            <Button
              id="btn-update-draft"
              variant="outline"
              onClick={() => handleSubmit()}
              disabled={isSubmitting}
              className="gap-2"
            >
              <Save className="h-4 w-4" />
              Save Draft
            </Button>
          )}
          <Button
            id="btn-update-publish"
            onClick={() => handleSubmit("published")}
            disabled={isSubmitting}
            className="gap-2"
          >
            <Send className="h-4 w-4" />
            {currentStatus === "published" ? "Update" : "Publish"}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-5">
          <div className="space-y-2">
            <Label htmlFor="edit-input-title">Title *</Label>
            <Input
              id="edit-input-title"
              placeholder="Enter a compelling title..."
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="text-lg font-medium h-12"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-input-excerpt">Excerpt</Label>
            <Textarea
              id="edit-input-excerpt"
              placeholder="A short summary shown in post cards (max 500 chars)..."
              value={excerpt}
              onChange={(e) => setExcerpt(e.target.value)}
              rows={3}
              maxLength={500}
              className="resize-none"
            />
            <p className="text-xs text-muted-foreground text-right">{excerpt.length}/500</p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-input-content">Content * (Markdown supported)</Label>
            <Textarea
              id="edit-input-content"
              placeholder="Write your post content here..."
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={20}
              className="font-mono text-sm resize-none"
            />
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-5">
          <div className="border rounded-xl p-4 bg-card space-y-5">
            <h3 className="font-semibold text-sm">Post Settings</h3>

            <div className="space-y-2">
              <Label htmlFor="edit-input-cover">Cover Image (Optional)</Label>
              <div className="flex gap-2">
                <Input
                  id="edit-input-cover"
                  placeholder="https://... or upload"
                  value={coverImage}
                  onChange={(e) => setCoverImage(e.target.value)}
                  disabled={isUploadingCover}
                  className="flex-1"
                />
                <div className="relative">
                  <Input
                    type="file"
                    accept="image/*"
                    onChange={handleCoverUpload}
                    disabled={isUploadingCover}
                    className="hidden"
                    id="cover-upload-file"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={() => document.getElementById("cover-upload-file")?.click()}
                    disabled={isUploadingCover}
                    title="Upload cover image"
                  >
                    {isUploadingCover ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Upload className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>
              {coverImage && (
                <div className="relative rounded-lg overflow-hidden h-28 bg-muted mt-2 group/cover">
                  <img
                    src={coverImage}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                    onError={(e) => {
                      ;(e.target as HTMLImageElement).style.display = "none"
                    }}
                  />
                  <Button
                    type="button"
                    variant="destructive"
                    size="icon"
                    className="absolute top-1.5 right-1.5 h-6 w-6 opacity-0 group-hover/cover:opacity-100 transition-opacity"
                    onClick={() => setCoverImage("")}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-select-category">Category</Label>
              <Select value={categoryId} onValueChange={(val) => setCategoryId(val ?? "")}>
                <SelectTrigger id="edit-select-category">
                  <SelectValue placeholder="Select a category" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">No category</SelectItem>
                  {categories.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Tags</Label>
              <div className="flex flex-wrap gap-2">
                {tags.map((t) => {
                  const selected = selectedTagIds.includes(t.id)
                  return (
                    <button
                      key={t.id}
                      id={`edit-tag-toggle-${t.id}`}
                      type="button"
                      onClick={() => toggleTag(t.id)}
                      className={`text-xs px-2.5 py-1 rounded-full border transition-colors cursor-pointer ${
                        selected
                          ? "bg-primary text-primary-foreground border-primary"
                          : "bg-transparent text-muted-foreground border-border hover:border-primary hover:text-foreground"
                      }`}
                    >
                      {t.name}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>

          {selectedTagIds.length > 0 && (
            <div className="border rounded-xl p-4 bg-card space-y-2">
              <h3 className="font-semibold text-sm">Selected Tags</h3>
              <div className="flex flex-wrap gap-1">
                {selectedTagIds.map((tagId) => {
                  const tag = tags.find((t) => t.id === tagId)
                  return tag ? (
                    <Badge
                      key={tagId}
                      variant="secondary"
                      className="gap-1 cursor-pointer"
                      onClick={() => toggleTag(tagId)}
                    >
                      {tag.name}
                      <X className="h-3 w-3" />
                    </Badge>
                  ) : null
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
